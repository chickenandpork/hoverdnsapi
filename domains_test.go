package hoverdnsapi_test

import (
	"fmt"
	"testing"

	"github.com/chickenandpork/hoverdnsapi" // to ensure testing without extra access
	json "github.com/gibson042/canonicaljson-go"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

const (
	// HoverAddressJson is used to check that certain domains' contact info is set yo Hover as
	// would be logical.  Recalling the reason for this structure involves Hover's tools not
	// allowing me to do bulk edits, I'm concerned that when my address changes, not all
	// domains get the change (if only it hadn't happened already).  Trust-by-verify, this
	// pertains to that second step.
	HoverAddressJson = `{"status":"active","org_name":"Hover, a service of Tucows.com Co","first_name":"Support","last_name":"Contact","address1":"96 Mowat Ave.","address2":"","address3":"","city":"Toronto","state":"ON","zip":"M6K 3M1","country":"CA","phone":"+1.8667316556","fax":"","email":"help@hover.com"}`
)

// mustJsonMarshal is a convenience function to make testcode a bit more streamlined with an inline
// conversion
func mustJsonMarshal(e interface{}) string {
	if data, err := json.MarshalIndent(e, "", "    "); err == nil {
		return string(data)
	}
	panic("Json parsing of a constant string implies a broken testdata")
}

// TestAPIURL covers wher I've screwed up the attempt to be simple with the API URL.  Seems I keep
// tracking two non-DRY equivalents.  Dammit, make them agree and simplify later.
func TestAPIURL(t *testing.T) {
	const thisWeeksURLToTry = "https://www.hover.com/api"

	var tests = []struct {
		desc     string
		code     string
		expected string
	}{
		{"login", "login", fmt.Sprintf("%s/%s", thisWeeksURLToTry, "login")},
		{"domains", "domains", fmt.Sprintf("%s/%s", thisWeeksURLToTry, "domains")},
		{"bogus", "otherwise", fmt.Sprintf("%s/%s", thisWeeksURLToTry, "otherwise")},
		{"domainpost", "domains/12345/dns", fmt.Sprintf("%s/%s", thisWeeksURLToTry, "domains/12345/dns")},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			observed := hoverdnsapi.APIURL(test.code)
			assert.Equal(t, test.expected, observed)
		})
	}
}

// TestParseAddress is a framework to test any problematic addresses; I've loaded it with just Hover's address for basic parse-testing
func TestParseAddress(t *testing.T) {
	var tests = []struct {
		desc     string
		jsonText string
		expected hoverdnsapi.Address
	}{
		// some buildup.  I had a strange error in parsing before.  I don't want it to sneak back in.
		{"Just A Status", `{"status": "active"}`, hoverdnsapi.Address{Status: "active"}},
		{"Just An Org", `{"org_name": "Lord of the Flies"}`, hoverdnsapi.Address{OrganizationName: "Lord of the Flies"}},
		{"Add An Org", `{"status": "active", "org_name": "Lord of the Flies"}`, hoverdnsapi.Address{Status: "active", OrganizationName: "Lord of the Flies"}},

		// effectively, this confirms the accuracy of a non-DRY copy between two constants
		{"Hover address", HoverAddressJson, hoverdnsapi.HoverAddress},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var observed hoverdnsapi.Address
			json.Unmarshal([]byte(test.jsonText), &observed)
			diff := deep.Equal(observed, test.expected)
			assert.Nilf(t, diff, `test "%s": expected "%+v" not matched by tested result "%+v", namely: %+v

Consider: %s`,
				test.desc, test.expected, observed, diff, mustJsonMarshal(test.expected))
		})
	}
}

// TestParseDomain test-parses the entire domain structure to ensure the json markup in the struct
// is accurate and functional.  Even trivial assumptions get broken, I'd rather catch them than
// merely assume.
func TestParseDomain(t *testing.T) {
	var tests = []struct {
		desc     string
		jsonText string
		expected hoverdnsapi.Domain
	}{
		// Although only one test, this is done as a an array of tests to permit trivial
		// extension as needed.  Add failing cases here, PR me fixes or ask me to merge
		// your broken case to a workspace and fix it for you.
		{"One Big Test", `{ "auto_renew": true, "id": "dom8675309", "domain_name": "chickenandpork.com", "num_emails": 0, "renewal_date": "2018-05-30", "display_date": "2018-05-30", "registered_date": "2000-05-30",  "contacts": {
        "admin": ` + HoverAddressJson + `,
        "owner": ` + HoverAddressJson + `,
        "tech": ` + HoverAddressJson + `,
        "billing": ` + HoverAddressJson + `}}`, hoverdnsapi.Domain{ID: "dom8675309", DomainName: "chickenandpork.com", AutoRenew: true, NumEmails: 0, RenewalDate: "2018-05-30", DisplayDate: "2018-05-30", RegisteredDate: "2000-05-30", Contacts: hoverdnsapi.ContactBlock{Admin: hoverdnsapi.HoverAddress, Billing: hoverdnsapi.HoverAddress, Owner: hoverdnsapi.HoverAddress, Tech: hoverdnsapi.HoverAddress}}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var observed hoverdnsapi.Domain
			json.Unmarshal([]byte(test.jsonText), &observed)
			diff := deep.Equal(observed, test.expected)
			assert.Nilf(t, diff, `test "%s": expected "%+v" not matched by tested result "%+v", namely: %+v

Consider: %s`,
				test.desc, test.expected, observed, diff, mustJsonMarshal(test.expected))
		})
	}
}

// TestRoundTrip ensures that parsing mapping has full coverage by taking a json sample, parsing to
// a Domain structure, and then back to json; if the starting json is of canonical form, then
// test-failures relate to missing parameters in the struct markup.  This allows me to cut-n-paste
// a verbatim json output, run it through this test, and catch changes in the upstream format.  I
// could concievably do this with a live query sample on a recurring basis.
func TestRoundTrip(t *testing.T) {
	sample := `    {
      "id": "dom8675309",
      "domain_name": "chickenandpork.com",
      "num_emails": 1,
      "renewal_date": "2020-05-30",
      "display_date": "2020-05-30",
      "registered_date": "2000-05-30",
      "status": "active",
      "auto_renew": true,
      "renewable": true,
      "locked": true,
      "whois_privacy": true,
      "nameservers": [
        "ns3.hover.com",
        "ns1.hover.com",
        "ns2.hover.com"
      ],
      "contacts": {
        "admin": {
          "first_name": "Allan",
          "org_name": "Chicken and Pork",
          "city": "Porkville",
          "country": "US",
          "status": "active",
          "email": "chickenandpork@example.com",
          "address3": "",
          "address1": "12345 SW Awesome St",
          "last_name": "Clark",
          "address2": "",
          "fax": "",
          "state": "PV",
          "phone": "+1.2505551291",
          "zip": "01337"
        },
        "owner": {
          "address1": "12345 SW Awesome St",
          "last_name": "Clark",
          "address3": "",
          "phone": "+1.2505551291",
          "state": "PV",
          "address2": "",
          "fax": "",
          "first_name": "Allan",
          "org_name": "Chicken and Pork",
          "city": "Porkville",
          "email": "chickenandpork@example.com",
          "country": "US",
          "status": "active",
          "zip": "01337"
        },
        "tech": {
          "phone": "+1.2505551291",
          "state": "PV",
          "fax": "",
          "address2": "",
          "address1": "12345 SW Awesome St",
          "last_name": "Clark",
          "address3": "",
          "email": "chickenandpork@example.com",
          "country": "US",
          "status": "active",
          "org_name": "Chicken and Pork",
          "first_name": "Allan",
          "city": "Porkville",
          "zip": "01337"
        },
        "billing": {
          "city": "Porkville",
          "org_name": "Chicken and Pork",
          "first_name": "Allan",
          "status": "active",
          "country": "US",
          "email": "chickenandpork@example.com",
          "address3": "",
          "last_name": "Clark",
          "address1": "12345 SW Awesome St",
          "state": "PV",
          "fax": "",
          "address2": "",
          "phone": "+1.2505551291",
          "zip": "01337"
        }
      },
      "glue": {},
      "hover_user": {
        "email": "chickenandpork@example.com",
        "email_secondary": "chickenandpork@example.com",
        "billing": {
          "pay_mode": "apple_pay",
          "description": "Black metal card ending 1337"
        }
      }
    }`

	canonizer := make(map[string]interface{}, 1)
	json.Unmarshal([]byte(sample), &canonizer)
	canonicalSample := mustJsonMarshal(canonizer)

	var observed hoverdnsapi.Domain
	json.Unmarshal([]byte(canonicalSample), &observed)

	observedJson := mustJsonMarshal(observed)

	assert.Equal(t, canonicalSample, observedJson)
}
