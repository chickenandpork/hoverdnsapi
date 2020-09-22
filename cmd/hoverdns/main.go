// CLI to do basic hoverdns create/update/delete
//
// (intentionally similar to: lego -a -m chickenandporn@gmail.com --domains smallfoot.org --dns hover run )
//
// For example:  (ie the testing I do when I manually test)
//     export PASSFILE=$(mktemp)
//     echo '{"username": "scott", "plaintextpassword": "tiger"}' > ${PASSFILE}
// NOTE: the env "PASSFILE" or "HOVER_PASSFILE" is looked for as a source of config in the test app
//
//     go build ./cmd/... && \
//     ./hoverdns -a -m "chickenandporn@gmail.com" -domains smallfoot.org -host test1 -value ABCDE add
// or:
//      go run ./cmd/... -a -m "chickenandporn@gmail.com" -domains smallfoot.org -host test1 -value ABCDE add
// (then, yeah, *sigh*, I manually check the domain:  dig -t TXT test1.smallfoot.org)
//
//     ./hoverdns -a -m "chickenandporn@gmail.com" -domains smallfoot.org -host test1 delete
// (similarly manuallt check that the record is gone, subject to TTL of 300 sec)

package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	hover "github.com/chickenandpork/hoverdnsapi"
	"github.com/go-acme/lego/v3/log"
	"github.com/urfave/cli/v2"
)

var version = "dev" // overwrite in production build if you ever want to ship this tool

var (
	onlyOneClient sync.Once
	client        *hover.Client
)

// getClient singletons a hover client
func getClient(username, password, passfile string) *hover.Client {
	onlyOneClient.Do(func() {
		fmt.Printf("logging in: u: %+v p: %+v f:%+v\n", username, password, passfile)
		client = hover.NewClient(username, password, passfile, 30*time.Second)
		client.FillDomains()
	})

	return client
}


func main() {
	var (
		passfile string
		password string
		username string
		domains  cli.StringSlice
		hostpart string
		value    string
		ttl      uint
	)

	app := &cli.App{
		Commands: []*cli.Command{
			// "info" dumps JSON of the remote DNS data to confirm it authenticates
			{Name: "info",
				Aliases: []string{"q", "check"},
				Usage:   "Check info about a domain; also confirms access",
				Action: func(c *cli.Context) error {
					fmt.Printf("info %v\n", domains.Value())
					if client := getClient(username, password, passfile); client != nil {
						for _, d := range domains.Value() {
							if do, ok := client.GetDomainByName(d); ok {
								fmt.Printf("Domain: %s ==> %#v", d, do)
							} else {
								fmt.Printf("Domain: %s not found\n", d)
							}
						}
					} else {
						fmt.Println("nope, Hover not instantiated")
						return fmt.Errorf("nope, Hover not instantiated")
					}

					return nil
				},
			},

			// "add" adds an action to the Actions stack for each domain, then executes, so
			// any dependencies such as expanding records don't need to leak out here: just
			// stack up the actions, let the subsys figure it out.
			{Name: "add",
				Usage: "add a record in each domain (extra parms for hostname and value)",
				Action: func(c *cli.Context) error {
					actions := make([]hover.Action, 0)
					fmt.Printf("adding to %v\n", domains.Value())
					for _, d := range domains.Value() {
						switch {
						case value == "":
						case hostpart == "":
						default:
							actions = append(actions, hover.NewAction(hover.Add, hostpart+"."+d, d, value, ttl))
						}
					}

					return getClient(username, password, passfile).DoActions(actions...)
				},
			},

			// "update" adds an Update action to the Actions stack for each domain, then
                        // executes, so any dependencies such as expanding records don't need to
			// leak out here: just stack up the actions, let the subsys figure it out.
			{Name: "update",
				Usage: "update a record in each domain (extra parms for hostname and value)",
				Action: func(c *cli.Context) error {
					actions := make([]hover.Action, 0)
					fmt.Printf("updating %v\n", domains.Value())
					for _, d := range domains.Value() {
						switch {
						case value == "":
						case hostpart == "":
						default:
							actions = append(actions, hover.NewAction(hover.Update, hostpart+"."+d, d, value, ttl))
						}
					}

					return getClient(username, password, passfile).DoActions(actions...)
				},
			},

			// "delete" adds a delete action to the Actions stack for each domain, then executes similar to
			// "add" above: any dependencies such as expanding records don't need to leak out here: just
			// stack up the actions, let the subsys figure it out.
			{Name: "delete",
				Usage: "delete a record in each domain (extra parm for hostname)",
				Action: func(c *cli.Context) error {
					actions := make([]hover.Action, 0)
					fmt.Printf("deleting from %v\n", domains.Value())
					for _, d := range domains.Value() {
						switch {
						case hostpart == "":
						default:
							actions = append(actions, hover.NewAction(hover.Delete, hostpart+"."+d, d, "", ttl))
						}
					}

					return getClient(username, password, passfile).DoActions(actions...)
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "accept-tos", Aliases: []string{"a"}, Usage: "placeholder to accept the current Let's Encrypt terms of service."},
			&cli.StringFlag{Name: "email", Aliases: []string{"m"}, Usage: "placeholder to accept Let's Encrypt account by email address", EnvVars: []string{"HOVER_EMAIL", "EMAIL"}},
			&cli.StringFlag{Name: "passfile", Usage: "username/password file", Destination: &passfile, EnvVars: []string{"HOVER_PASSFILE", "PASSFILE"}},
			&cli.StringFlag{Name: "password", Usage: "password if not using passfile", Destination: &password, EnvVars: []string{"HOVER_PASSWORD", "PASSWORD"}},
			&cli.StringFlag{Name: "username", Usage: "username if not using passfile", Destination: &username, EnvVars: []string{"HOVER_USERNAME", "USERNAME"}},
			&cli.StringSliceFlag{Name: "domains", Usage: "domain(s) to act upon", Destination: &domains, EnvVars: []string{"HOVER_DOMAINS", "DOMAINS"}},
			&cli.StringFlag{Name: "host", Usage: `relative hostname  added/deleted (not FQDN, but the host in "host.${domain}")`, Destination: &hostpart},
			&cli.StringFlag{Name: "value", Usage: "DNS Value (ie TXT record value)", Destination: &value},
			&cli.UintFlag{Name: "ttl", Usage: "TTL of zone value if added", Value: 300, Destination: &ttl},
		},
		HelpName: "hoverdns",
		Name:     "hoverdns",
		Usage:    "Hover DNS CLI Client",
		//Action: func(c *cli.Context) error {
		//	fmt.Println("That command doesn't seem to be a functional action to perform")
		//	return nil
		//},
		Version: version,
	}

	app.EnableBashCompletion = true
	// app.Setup() // run later in app.Run() but sentinels to see if it's already run, like run here

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version %s %s/%s\n", app.Name, app.Version, runtime.GOOS, runtime.GOARCH)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
