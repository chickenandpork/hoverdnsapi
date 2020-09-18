package hoverdnsapi_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

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

// TestAPIURL covers where I've screwed up the attempt to be simple with the API URL.  Seems I keep
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

func TestParseDomainBigResponse(t *testing.T) {
	var response hoverdnsapi.DomainList

	bigResponse := `{"succeeded":true,"domains":[{"id":"dom481005","domain_name":"chickenandpork.com","num_emails":0,"renewal_date":"2021-01-28","display_date":"2021-01-28","registered_date":"2010-01-28","status":"active","auto_renew":true,"renewable":true,"locked":true,"whois_privacy":true,"nameservers":["ns1.hover.com","ns2.hover.com"],"contacts":{"admin":{"email":"chickenandporn@gmail.com","state":"PV","address3":"","city":"Porkville","org_name":"n/a","country":"US","phone":"+1.2505551291","fax":"","last_name":"Clark","first_name":"Allan","address1":"12345 Awesome St","status":"active","address2":"","zip":"01337"},"owner":{"address3":"","org_name":"n/a","city":"Porkville","email":"chickenandporn@gmail.com","state":"PV","phone":"+1.2505551291","country":"US","last_name":"Clark","fax":"","status":"active","address2":"","address1":"12345 Awesome St","first_name":"Allan","zip":"01337"},"tech":{"city":"Porkville","org_name":"n/a","address3":"","state":"PV","email":"chickenandporn@gmail.com","phone":"+1.2505551291","country":"US","last_name":"Clark","fax":"","address2":"","status":"active","first_name":"Allan","address1":"12345 Awesome St","zip":"01337"},"billing":{"phone":"+1.2505551291","country":"US","address3":"","city":"Porkville","org_name":"n/a","email":"chickenandporn@gmail.com","state":"PV","address2":"","status":"active","first_name":"Allan","address1":"12345 Awesome St","last_name":"Clark","fax":"","zip":"01337"}},"glue":{},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}}},{"id":"dom727321","domain_name":"chickenandporn.com","num_emails":1,"renewal_date":"2020-05-30","display_date":"2020-05-30","registered_date":"2000-05-30","status":"active","auto_renew":true,"renewable":true,"locked":true,"whois_privacy":true,"nameservers":["ns3.hover.com","ns1.hover.com","ns2.hover.com"],"contacts":{"admin":{"first_name":"Allan","org_name":"Chicken and Pork","city":"Porkville","country":"US","status":"active","email":"chickenandporn@gmail.com","address3":"","address1":"12345 Awesome St","last_name":"Clark","address2":"","fax":"","state":"PV","phone":"+1.2505551291","zip":"01337"},"owner":{"address1":"12345 Awesome St","last_name":"Clark","address3":"","phone":"+1.2505551291","state":"PV","address2":"","fax":"","first_name":"Allan","org_name":"Chicken and Pork","city":"Porkville","email":"chickenandporn@gmail.com","country":"US","status":"active","zip":"01337"},"tech":{"phone":"+1.2505551291","state":"PV","fax":"","address2":"","address1":"12345 Awesome St","last_name":"Clark","address3":"","email":"chickenandporn@gmail.com","country":"US","status":"active","org_name":"Chicken and Pork","first_name":"Allan","city":"Porkville","zip":"01337"},"billing":{"city":"Porkville","org_name":"Chicken and Pork","first_name":"Allan","status":"active","country":"US","email":"chickenandporn@gmail.com","address3":"","last_name":"Clark","address1":"12345 Awesome St","state":"PV","fax":"","address2":"","phone":"+1.2505551291","zip":"01337"}},"glue":{},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}}},{"id":"dom2719843","domain_name":"fsme.io","num_emails":0,"renewal_date":"2020-07-08","display_date":"2020-07-08","registered_date":"2019-07-08","status":"active","auto_renew":true,"renewable":true,"locked":true,"whois_privacy":"unsupported","nameservers":["ns1.hover.com","ns2.hover.com"],"contacts":{"admin":{"status":"active","address2":"","first_name":"Allan","address1":"291 Westgate Rd","last_name":"Clark","fax":"","phone":"+1.2505551291","country":"CA","city":"Campbell River","org_name":"Allan Clark","address3":"","state":"BC","email":"chickenandporn@gmail.com","zip":"V9W1R7"},"owner":{"last_name":"Clark","fax":"","address2":"","status":"active","address1":"291 Westgate Rd","first_name":"Allan","city":"Campbell River","org_name":"Allan Clark","address3":"","state":"BC","email":"chickenandporn@gmail.com","country":"CA","phone":"+1.2505551291","zip":"V9W1R7"},"tech":{"country":"CA","phone":"+1.8667316556","state":"ON","email":"help@hover.com","org_name":"Hover, a service of Tucows.com Co","city":"Toronto","address3":"","first_name":"Support","address1":"96 Mowat Ave.","address2":"","status":"active","fax":"","last_name":"Contact","zip":"M6K 3M1"},"billing":{"fax":"","last_name":"Clark","first_name":"Allan","address1":"291 Westgate Rd","status":"active","address2":"","email":"chickenandporn@gmail.com","state":"BC","address3":"","org_name":"Allan Clark","city":"Campbell River","country":"CA","phone":"+1.2505551291","zip":"V9W1R7"}},"glue":{},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}}},{"id":"dom2907388","domain_name":"minutix.com","num_emails":0,"renewal_date":"2020-12-11","display_date":"2020-12-11","registered_date":"2019-12-11","status":"active","auto_renew":true,"renewable":true,"locked":true,"whois_privacy":true,"nameservers":["ns1.hover.com","ns2.hover.com"],"contacts":{"admin":{"email":"chickenandporn@gmail.com","country":"US","status":"active","first_name":"Allan","org_name":"n/a","city":"Porkville","phone":"+1.2505551291","address2":"","fax":"","state":"PV","address1":"12345 Awesome St","last_name":"Clark","address3":"","zip":"01337"},"owner":{"email":"chickenandporn@gmail.com","country":"US","status":"active","first_name":"Allan","org_name":"n/a","city":"Porkville","phone":"+1.2505551291","state":"PV","address2":"","fax":"","address1":"12345 Awesome St","last_name":"Clark","address3":"","zip":"01337"},"tech":{"phone":"+1.4165385498","state":"ON","address2":"","fax":"","address1":"96 Mowat Avenue","last_name":"Contact","address3":"","email":"help@hover.com","country":"CA","status":"active","org_name":"Hover, a service of Tucows Inc.","first_name":"Technical","city":"Toronto","zip":"M6K 3M1"},"billing":{"first_name":"Allan","org_name":"n/a","city":"Porkville","country":"US","status":"active","email":"chickenandporn@gmail.com","address3":"","address1":"12345 Awesome St","last_name":"Clark","address2":"","fax":"","state":"PV","phone":"+1.2505551291","zip":"01337"}},"glue":{},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}}},{"id":"dom675971","domain_name":"qloak.me","num_emails":0,"renewal_date":"2021-01-01","display_date":"2021-01-01","registered_date":"2012-01-01","status":"active","auto_renew":true,"renewable":true,"locked":true,"whois_privacy":true,"nameservers":["ns1.hover.com","ns2.hover.com"],"contacts":{"admin":{"status":"active","address3":"","org_name":"n/a","last_name":"Clark","city":"Porkville","phone":"+1.2505551291","state":"PV","address1":"12345 Awesome St","country":"US","first_name":"Allan","email":"chickenandporn@gmail.com","address2":"","fax":"","zip":"01337"},"owner":{"city":"Porkville","address1":"12345 Awesome St","state":"PV","phone":"+1.2505551291","org_name":"n/a","address3":"","status":"active","last_name":"Clark","address2":"","fax":"","country":"US","email":"chickenandporn@gmail.com","first_name":"Allan","zip":"01337"},"tech":{"address2":"","fax":"","country":"CA","first_name":"Technical","email":"help@hover.com","city":"Toronto","phone":"+1.4165385498","state":"ON","address1":"96 Mowat Avenue","status":"active","address3":"","org_name":"Hover, a service of Tucows Inc.","last_name":"Contact","zip":"M6K 3M1"},"billing":{"fax":"","address2":"","email":"chickenandporn@gmail.com","first_name":"Allan","country":"US","address1":"12345 Awesome St","state":"PV","phone":"+1.2505551291","city":"Porkville","last_name":"Clark","org_name":"n/a","status":"active","address3":"","zip":"01337"}},"glue":{},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}}},{"id":"dom202932","domain_name":"secret-island-lair.ca","num_emails":1,"renewal_date":"2020-11-12","display_date":"2020-11-12","registered_date":"2007-11-12","status":"active","auto_renew":false,"renewable":true,"locked":true,"whois_privacy":true,"nameservers":["ns1.domaindirect.com","ns2.domaindirect.com","ns3.domaindirect.com"],"contacts":{"admin":{"address1":"291 Westgate Rd","last_name":"Clark","address3":"","phone":"+1.2505551291","address2":"","fax":"","state":"BC","first_name":"Allan","org_name":"Allan Clark","city":"Campbell River","lang":"","email":"chickenandporn@gmail.com","country":"CA","status":"active","zip":"V9W1R7"},"owner":{"address2":"","state":"BC","fax":"","phone":"+1.2505551291","address3":"","address1":"291 Westgate Rd","last_name":"Clark","country":"CA","status":"active","email":"chickenandporn@gmail.com","lang":"","first_name":"Allan","org_name":"Allan Clark","city":"Campbell River","zip":"V9W1R7"},"tech":{"status":"active","country":"CA","email":"help@hover.com","lang":"","city":"Toronto","first_name":"Support","org_name":"Hover, a service of Tucows.com Co","fax":"","state":"ON","address2":"","phone":"+1.8667316556","address3":"","last_name":"Contact","address1":"96 Mowat Ave.","zip":"M6K3M1"}},"glue":{},"registry_settings":{"type":"CA","legal_type":"CCT"},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}},"readonly_attributes":["contacts.owner.country","contacts.owner.first_name","contacts.owner.last_name"]},{"id":"dom204428","domain_name":"secret-island-lair.com","num_emails":1,"renewal_date":"2020-11-12","display_date":"2020-11-12","registered_date":"2007-11-12","status":"active","auto_renew":false,"renewable":true,"locked":false,"whois_privacy":true,"nameservers":["ns1.domaindirect.com","ns2.domaindirect.com","ns3.domaindirect.com"],"contacts":{"admin":{"phone":"+1.2505551291","address2":"","state":"PV","fax":"","last_name":"Clark","address1":"12345 Awesome St","address3":"","email":"chickenandporn@gmail.com","status":"active","country":"US","city":"Porkville","first_name":"Allan","org_name":"n/a","zip":"01337"},"owner":{"address1":"12345 Awesome St","last_name":"Clark","address3":"","phone":"+1.2505551291","address2":"","state":"PV","fax":"","first_name":"Allan","org_name":"n/a","city":"Porkville","email":"chickenandporn@gmail.com","country":"US","status":"active","zip":"01337"},"tech":{"address3":"","address1":"96 Mowat Avenue","last_name":"Contact","state":"ON","address2":"","fax":"","phone":"+1.4165385498","first_name":"Technical","org_name":"Hover, a service of Tucows Inc.","city":"Toronto","country":"CA","status":"active","email":"help@hover.com","zip":"M6K 3M1"},"billing":{"fax":"","address2":"","state":"PV","phone":"+1.2505551291","address3":"","address1":"12345 Awesome St","last_name":"Clark","country":"US","status":"active","email":"chickenandporn@gmail.com","first_name":"Allan","org_name":"n/a","city":"Porkville","zip":"01337"}},"glue":{},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}}},{"id":"dom202730","domain_name":"secretislandlair.ca","num_emails":1,"renewal_date":"2020-11-12","display_date":"2020-11-12","registered_date":"2007-11-12","status":"active","auto_renew":false,"renewable":true,"locked":true,"whois_privacy":true,"nameservers":["ns1.domaindirect.com","ns2.domaindirect.com","ns3.domaindirect.com"],"contacts":{"admin":{"address3":"","address1":"291 Westgate Rd","org_name":"Allan Clark","last_name":"Clark","city":"Campbell River","state":"BC","phone":"+1.2505551291","first_name":"Allan","lang":"","status":"active","fax":"","address2":"","country":"CA","email":"chickenandporn@gmail.com","zip":"V9W1R7"},"owner":{"org_name":"Allan Clark","last_name":"Clark","city":"Campbell River","address1":"291 Westgate Rd","address3":"","first_name":"Allan","state":"BC","phone":"+1.2505551291","fax":"","status":"active","lang":"","address2":"","country":"CA","email":"chickenandporn@gmail.com","zip":"V9W1R7"},"tech":{"city":"Toronto","org_name":"Hover, a service of Tucows.com Co","last_name":"Contact","address3":"","address1":"96 Mowat Ave.","first_name":"Support","state":"ON","phone":"+1.8667316556","fax":"","lang":"","status":"active","email":"help@hover.com","address2":"","country":"CA","zip":"M6K3M1"}},"glue":{},"registry_settings":{"type":"CA","legal_type":"CCT"},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}},"readonly_attributes":["contacts.owner.country","contacts.owner.first_name","contacts.owner.last_name"]},{"id":"dom203659","domain_name":"secretislandlair.com","num_emails":1,"renewal_date":"2020-11-12","display_date":"2020-11-12","registered_date":"2007-11-12","status":"active","auto_renew":false,"renewable":true,"locked":true,"whois_privacy":true,"nameservers":["ns1.hover.com","ns2.hover.com"],"contacts":{"admin":{"state":"PV","address2":"","fax":"","phone":"+1.2505551291","address3":"","last_name":"Clark","address1":"12345 Awesome St","status":"active","country":"US","email":"chickenandporn@gmail.com","city":"Porkville","org_name":"n/a","first_name":"Allan","zip":"01337"},"owner":{"first_name":"Allan","org_name":"n/a","city":"Porkville","country":"US","status":"active","email":"chickenandporn@gmail.com","address3":"","address1":"12345 Awesome St","last_name":"Clark","state":"PV","address2":"","fax":"","phone":"+1.2505551291","zip":"01337"},"tech":{"org_name":"Hover, a service of Tucows Inc.","first_name":"Technical","city":"Toronto","email":"help@hover.com","country":"CA","status":"active","address1":"96 Mowat Avenue","last_name":"Contact","address3":"","phone":"+1.4165385498","state":"ON","address2":"","fax":"","zip":"M6K 3M1"},"billing":{"address3":"","last_name":"Clark","address1":"12345 Awesome St","state":"PV","fax":"","address2":"","phone":"+1.2505551291","city":"Porkville","first_name":"Allan","org_name":"n/a","status":"active","country":"US","email":"chickenandporn@gmail.com","zip":"01337"}},"glue":{},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}}},{"id":"dom202142","domain_name":"smallfoot.org","num_emails":0,"renewal_date":"2021-03-21","display_date":"2021-03-21","registered_date":"2003-03-21","status":"active","auto_renew":true,"renewable":true,"locked":true,"whois_privacy":true,"nameservers":["ns1.domaindirect.com","ns2.domaindirect.com","ns3.domaindirect.com"],"contacts":{"admin":{"city":"Porkville","last_name":"Clark","org_name":"n/a","address3":"","address1":"12345 Awesome St","first_name":"Allan","state":"PV","phone":"+1.2505551291","fax":"","status":"active","email":"chickenandporn@gmail.com","address2":"","country":"US","zip":"01337"},"owner":{"address1":"12345 Awesome St","address3":"","last_name":"Clark","org_name":"n/a","city":"Porkville","state":"PV","phone":"+1.2505551291","first_name":"Allan","status":"active","fax":"","address2":"","country":"US","email":"chickenandporn@gmail.com","zip":"01337"},"tech":{"status":"active","fax":"","address2":"","country":"US","email":"chickenandporn@gmail.com","address1":"12345 Awesome St","address3":"","org_name":"n/a","last_name":"Clark","city":"Porkville","state":"PV","phone":"+1.2505551291","first_name":"Allan","zip":"01337"},"billing":{"email":"chickenandporn@gmail.com","country":"US","address2":"","status":"active","fax":"","state":"PV","phone":"+1.2505551291","first_name":"Allan","address1":"12345 Awesome St","address3":"","city":"Porkville","org_name":"n/a","last_name":"Clark","zip":"01337"}},"glue":{},"hover_user":{"email":"chickenandporn@gmail.com","email_secondary":"chickenandporn@gmail.com","billing":{"pay_mode":"apple_pay","description":"Visa ending in 1234"}}}]}`
	json.Unmarshal([]byte(bigResponse), &response)

	assert.Greater(t, len(response.Domains), 1)
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

// TestEntriesRoundTrip builds upon TestRoundTrip to ensures that parsing mapping has full coverage
// including Domains expanded with entries by taking a json sample, parsing to a Domain structure,
// and then back to json; if the starting json is of canonical form, then test-failures relate to
// missing parameters in the struct markup.  This allows me to cut-n-paste a verbatim json output,
// run it through this test, and catch changes in the upstream format.  I could concievably do this
// with a live query sample on a recurring basis.
//
// This test case should probably be combined with the preceding
func TestEntriesRoundTrip(t *testing.T) {
	sample := `{
  "succeeded": true,
  "domains": [
    {

      "contacts": {
        "admin": { "address1": "", "address2": "", "address3": "", "city": "", "country": "", "email": "", "fax": "", "first_name": "", "last_name": "", "org_name": "", "phone": "", "state": "", "status": "", "zip": "" },
        "billing": { "address1": "", "address2": "", "address3": "", "city": "", "country": "", "email": "", "fax": "", "first_name": "", "last_name": "", "org_name": "", "phone": "", "state": "", "status": "", "zip": "" },
        "owner": { "address1": "", "address2": "", "address3": "", "city": "", "country": "", "email": "", "fax": "", "first_name": "", "last_name": "", "org_name": "", "phone": "", "state": "", "status": "", "zip": "" },
        "tech": { "address1": "", "address2": "", "address3": "", "city": "", "country": "", "email": "", "fax": "", "first_name": "", "last_name": "", "org_name": "", "phone": "", "state": "", "status": "", "zip": "" }
      },
      "display_date": "",
      "glue": {}, "hover_user": { "billing": {} },


      "domain_name": "secretislandlair.ca",
      "id": "dom202730",
      "active": true,
      "entries": [
        {
          "id": "dns1374387",
          "name": "@",
          "type": "A",
          "content": "64.98.145.30",
          "ttl": 900,
          "is_default": true,
          "can_revert": false
        },
        {
          "id": "dns1374388",
          "name": "*",
          "type": "A",
          "content": "64.98.145.30",
          "ttl": 900,
          "is_default": true,
          "can_revert": false
        },
        {
          "id": "dns1374389",
          "name": "www",
          "type": "A",
          "content": "64.98.145.30",
          "ttl": 900,
          "is_default": false,
          "can_revert": false
        },
        {
          "id": "dns1374390",
          "name": "smtp",
          "type": "CNAME",
          "content": "smtp.secretislandlair.ca.cust.hostedemail.com",
          "ttl": 900,
          "is_default": false,
          "can_revert": false
        },
        {
          "id": "dns1374391",
          "name": "mail",
          "type": "CNAME",
          "content": "mail.secretislandlair.ca.cust.hostedemail.com",
          "ttl": 900,
          "is_default": false,
          "can_revert": true
        },
        {
          "id": "dns1374392",
          "name": "@",
          "type": "MX",
          "content": "10 mx.secretislandlair.ca.cust.hostedemail.com",
          "ttl": 900,
          "is_default": false,
          "can_revert": true
        }
      ]
    }
  ]
}
`
	canonizer := make(map[string]interface{}, 1)
	json.Unmarshal([]byte(sample), &canonizer)
	canonicalSample := mustJsonMarshal(canonizer)

	var observed hoverdnsapi.DomainList
	json.Unmarshal([]byte(canonicalSample), &observed)

	observedJson := mustJsonMarshal(observed)

	assert.Equal(t, canonicalSample, observedJson)
}

// TestNewClientPasswordFile is similar to passfile_test::TestParseJsonFile which leverages
// ReadConfigFile, but in this case the test is one step shallower on the callstack: the part of
// creating a client connection from available information such as a given user/pass, or the name
// of a file on-disk.
//
// I intend to test both calls: giving the user/pass directly, and via a filename, yo both confirm
// that the results are similar and that they both work.  If I can elevate this to use the same
// test content, then the tests at this level should automatically cover tests at the lowerlevel,
// when extended.
func TestNewClient(t *testing.T) {
	for _, tt := range BasicTests {
		t.Run(tt.Name, func(t *testing.T) {
			tmpfile, err := ioutil.TempFile("", "testfile")
			if assert.NoErrorf(t, err, "Error creating temp file: %s", "formatted") {
				defer os.Remove(tmpfile.Name())

				_, err = tmpfile.Write([]byte(tt.Config))
				if assert.NoErrorf(t, err, "Error writing temp file: %s", "formatted") {
					observed, err := hoverdnsapi.ReadConfigFile(tmpfile.Name())
					if assert.NoErrorf(t, err, "ReadConfigFile returned an error: %s", "formatted") {
						if diff := deep.Equal(tt.Expected, *observed); diff != nil {
							t.Error(diff)
						}
					}

					client := hoverdnsapi.NewClient(observed.Username, observed.PlaintextPassword, "", 90*time.Second)
					assert.Equal(t, observed.Username, client.Username)
					assert.Equal(t, observed.PlaintextPassword, client.Password)

					fclient := hoverdnsapi.NewClient("", "", tmpfile.Name(), 90*time.Second)
					assert.Equal(t, observed.Username, fclient.Username)
					assert.Equal(t, observed.PlaintextPassword, fclient.Password)
				}
			}
		})
	}
}
