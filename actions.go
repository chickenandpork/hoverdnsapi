package hoverdnsapi

import (
	"fmt"
	"net/url"
)

// HoverAct is simply an enum type to typecheck various actions we can perform in a queue
type HoverAct int

const (
	// Add a domain record
	Add HoverAct = iota
	// Delete of a record will require the list of records
	Delete
	// Update will also need the list of records and IDs
	Update
)

// String of course gives a string representation of the Act code
func (h HoverAct) String() string {
	switch h {
	case Add:
		return "Add"
	case Delete:
		return "Delete"
	case Update:
		return "Update"
	}

	return "(error) HoverAct const extended without String() equivalent"
}

//const authHeader = "hoverauth"

// Action is a single action (Add, Update, Delete) to complete in a DoActions() call
type Action struct {
	action HoverAct
	fqdn   string
	domain string
	value  string
	ttl    int
}

func (a Action) String() string {
	return fmt.Sprintf("{action:%s fqdn:%s domain:%s, value:%s ttl:%d}", a.action, a.domain, a.fqdn, a.value, a.ttl)
}

// DoActions is a way to burn down an accumulated list of actions.  Mostly, this stack will be one
// or two deep, but this offers the chance to go grab a GetAuth() (for authentication cookie) if
// needed, or a detailed DNS list if needed, in a sort of lazy-evaluation logic that avoid these
// actions if not needed.
func (c *Client) DoActions(actions ...Action) (err error) {
	if len(c.domains.Domains) < 1 {
		if err = c.FillDomains(); err != nil {
			return err
		}
	}

	// {
	//     action:2 fqdn:_acme-challenge.domain.com domain: domain.com
	//     value:xzLAGicQ1PtUwmXLyCsagNI7O4m_Zsn8mcVREy7QrfY ttl:3600
	// }
	for actnum, a := range actions {
		if domain, ok := c.GetDomainByName(a.domain); ok {
			switch a.action {
			case Add:
				if resp, err := c.HTTPClient.PostForm(APIURLDNS(domain.ID), url.Values{
					"name":    {a.fqdn},
					"type":    {"TXT"},
					"content": {a.value},
				}); err != nil {
					fmt.Printf("hover: Info: posting threw: [%+v]\n", err)
				} else {
					resp.Body.Close()
				}
				//case Delete:
				//case Update:
			}
			fmt.Printf("Action Stack (%02d): [%s]\n", actnum, a)
		} else {
			c.log.Printf("domain %s not found", a.domain)
			return fmt.Errorf("Domain %s not found in domains", a.domain)
		}
	}
	return nil
}
