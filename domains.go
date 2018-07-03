package hoverdnsapi

// Address holds an address used for admin, billing, or tech contact.  Empirically, it seems at
// least US and Canada formats are squeezed into a US format.  Please PR if you discover additional
// formats.
type Address struct {
	Status           string `json:"status"`     // Status seems to be "active" in all my zones
	OrganizationName string `json:"org_name"`   // Name of Organization
	FirstName        string `json:"first_name"` // First naem seems to be given non-family name, not positional
	LastName         string `json:"last_name"`  // Last Name seems to be family name, not positional
	Address1         string `json:"address1"`
	Address2         string `json:"address2"`
	Address3         string `json:"address3"`
	City             string `json:"city"`
	State            string `json:"state"`   // State seems to be the US state or the Canadian province
	Zip              string `json:"zip"`     // 5-digit US (ie 10001) or 6-char slammed Canadian (V0H1X0 no space)
	Country          string `json:"country"` // 2-leter state code; this seems to match the second (non-country) of a ISO-3166-2 code
	Phone            string `json:"phone"`   // phone format all over the map, but thy seem to write it as a ITU E164, but a "." separating country code and subscriber number
	Facsimile        string `json:"fax"`     // same format as phone
	Email            string `json:"email"`   // rfc2822 format email address such as rfc2822 para 3.4.1
}

// ContactBlock is merely the four contact addresses that Hover uses, but it's easier to work with
// a defined type in static constants during testing
type ContactBlock struct {
	Admin   Address `json:"admin"`
	Billing Address `json:"billing"`
	Tech    Address `json:"tech"`
	Owner   Address `json:"owner"`
}

// Domain structure describes the config for an entire domain within Hover: the dates involved,
// contact addresses, nameservers, etc: it seems to cover everything about the domain in one
// structure, which is convenient when you want to compare data across many domains.
type Domain struct {
	ID             string       `json:"id"`              // A unique opaque identifier defined by Hover
	DomainName     string       `json:"domain_name"`     // the actual domain name.  ie: "example.com"
	NumEmails      int          `json:"num_emails"`      // This appears to be the number of email accounts either permitted or defined for the domain
	RenewalDate    string       `json:"renewal_date"`    // This renewal date appears to be the first day of non-service after a purchased year of valid service: the first day offline if you don't renew.  RFC3339/ISO8601 -formatted yyyy-mm-dd.
	DisplayDate    string       `json:"display_date"`    // Display Date seems to be the same as Renewal Date but perhaps can allow for odd display corner-cases such as leap-years, leap-seconds, or timezones oddities.  RFC3339/ISO8601 to granularity of day as well.
	RegisteredDate string       `json:"registered_date"` // Date the domain was first registered, which is likely also the first day of service (or partial-day, technically)  RFC3339/ISO8601 to granularity of day as well.
	Contacts       ContactBlock `json:"contacts"`
	HoverUser      User         `json:"hover_user"`
	Glue           struct{}     `json:"glue"` // I'm not sure how Hover records Glue Records here, or whether they're still used.  Please PR a suggested format!
	NameServers    []string     `json:"nameservers"`
	Locked         bool         `json:"locked"`
	Renewable      bool         `json:"renewable"`
	AutoRenew      bool         `json:"auto_renew"`
	Status         string       `json:"status"`        // Status seems to be "active" in all my zones
	WhoisPrivacy   bool         `json:"whois_privacy"` // boolean as lower-case string: keep your real address out of whois?
}

// The User record in a Domain seems to record additional contact information that augments the
// Billing Contact with the credit card used and some metadata around it.
type User struct {
	Billing struct {
		Description string `json:"description"` // This seems to be a descirption of my card, such as "Visa ending 1234"
		PayMode     string `json:"pay_mode"`    // some reference to how payments are processed:  mine all say "apple_pay", and they're in my Apple Pay Wallet, but my account on Hover predates the existence of Apple Wallet, so ... I'm not sure
	} `json:"billing"`
	Email          string `json:"email"`
	EmailSecondary string `json:"email_secondary"`
}

var (
	// HoverAddress is a constant-ish var that I use to ensure that within my domains, the ones
	// I expect to have Hovers contact info (their default) do.  For example, Tech Contacts
	// where I don't want to be that guy (for managed domains, they should be the tech
	// contact).  Of course, if the values in this constant are incorrect, TuCows is the
	// authority, but please PR me a correction to help me maintain accuracy.
	HoverAddress = Address{
		Status:           "active",
		OrganizationName: "Hover, a service of Tucows.com Co",
		FirstName:        "Support",
		LastName:         "Contact",
		Address1:         "96 Mowat Ave.",
		City:             "Toronto",
		State:            "ON",
		Zip:              "M6K 3M1",
		Country:          "CA",
		Phone:            "+1.8667316556",
		Email:            "help@hover.com",
	}
)
