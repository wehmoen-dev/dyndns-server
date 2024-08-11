package main

import (
	"dyndns/pkg/client"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {

	err := (&cli.App{
		Name:                 "dyndns-client",
		Description:          "Client for Dynamic DNS",
		Usage:                "Client for Dynamic DNS",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "list-providers",
				Description: "List available IP providers to be used for grabbing the IP address. ",
				Usage:       "List available IP providers to be used for grabbing the IP address. ",
				Action: func(context *cli.Context) error {
					grabber := client.NewIpGrabber(nil)
					providers := grabber.GetAvailableHosts()

					t := table.NewWriter()
					t.SetStyle(table.StyleColoredBlackOnBlueWhite)
					t.SetOutputMirror(os.Stdout)
					t.AppendHeader(table.Row{"Provider", "IPV4 Host", "IPV6 Host"})

					for i, provider := range providers {
						hosts := client.IPGrabberOptions[provider]

						if i == 0 {
							t.AppendRow(table.Row{fmt.Sprintf("%s (Default)", provider), hosts.IPv4Host, hosts.IPv6Host})

						} else {
							t.AppendRow(table.Row{provider, hosts.IPv4Host, hosts.IPv6Host})
						}
					}

					t.Render()

					log.Printf("[DynDNS Client] Usage: dyndns-client update --host <host> [--username <username>] [--password <password>] [--ip-provider <ip-provider>]")

					return nil
				},
			},
			{
				Name:        "update",
				Args:        true,
				ArgsUsage:   `-- --server-url <server-url> [--username <username>] [--password <password>] [--ip-provider <ip-provider>]`,
				Description: "Grabs the IP address and updates the DNS records",
				Usage:       "Grabs the IP address and updates the DNS records",
				Action: func(context *cli.Context) error {

					args := client.ParseArgs(context.Args().Slice())

					host, err := args.Get("server-url")

					if err != nil {
						if errors.Is(err, client.ArgNotFoundError) {
							log.Printf("[DynDNS Client] Error: Server url not found. Please specify your server url by using the --server-url flag")
							return nil
						}
					}

					useAuth := true
					var auth string

					username, err := args.Get("username")
					if err != nil {
						useAuth = false
					}
					password, err := args.Get("password")
					if err != nil {
						useAuth = false
					}

					if useAuth {
						auth = fmt.Sprintf("%s:%s", username, password)
					} else {
						auth = ""
					}

					ipGrabber := client.NewIpGrabber(nil)

					grabberHostname, err := args.Get("ip-provider")

					if err != nil {
						if errors.Is(err, client.ArgNotFoundError) {
							log.Printf("[DynDNS Client] Error: ip-provider not found. Please specify your ip-provider by using the --ip-provider flag. Using default provider.")
							grabberHostname = "icanhazipcom"
						}
					}

					grabberHosts := client.IPGrabberOptions[grabberHostname]

					if grabberHosts == nil {
						log.Printf("[DynDNS Client] Error: ip-provider \"%s\" not found. Please specify your ip-provider by using the --ip-provider flag. Using default provider.", grabberHostname)
						grabberHosts = client.IPGrabberOptions["icanhazipcom"]
					}

					ipGrabber.SetHosts(grabberHosts)

					v4Address, err := ipGrabber.GrabV4()

					if err != nil {
						log.Printf("[DynDNS Client] Failed to retrieve IPv4 address: %v", err)
					}

					v6Address, err := ipGrabber.GrabV6()

					if err != nil {
						log.Printf("[DynDNS Client] Failed to retrieve IPv6 address: %v", err)
					}

					caller := client.NewRemoteApiCaller()

					result, err := caller.Call(host, v4Address, v6Address, auth)

					if err != nil {

						if errors.Is(err, client.ErrUnauthorized) {
							log.Printf("[DynDNS Client] Unauthorized. Please check your credentials.")
						}

						if errors.Is(err, client.ErrInternalServerError) {
							log.Printf("[DynDNS Client] Internal server error. Please try again later.")
						}

						if !errors.Is(err, client.ErrUnauthorized) && !errors.Is(err, client.ErrInternalServerError) {
							log.Printf("[DynDNS Client] Error calling server: %v", err)
						}
						return nil
					}

					if result == nil && err == nil {
						log.Printf("[DynDNS Client] Error calling server: result is nil")
						return errors.New("result is nil")
					}

					log.Printf("[DynDNS Client] Successfully updated DNS records for: %s", result.Name)

					if v4Address != nil {
						if result.V4.Success {
							if result.V4.Created {
								log.Printf("[DynDNS Client] Successfully created IPv4 record: %s", v4Address.String())
							} else {
								log.Printf("[DynDNS Client] Successfully updated IPv4 record: %s", v4Address.String())
							}
						} else {
							log.Printf("[DynDNS Client] Failed to update IPv4 record: %s. Error: %s", v4Address.String(), result.V4.Error)
						}
					} else {
						log.Printf("[DynDNS Client] No IPv4 address provided. Skipping IPv4 update.")
					}

					if v6Address != nil {
						if result.V6.Success {
							if result.V6.Created {
								log.Printf("[DynDNS Client] Successfully created IPv6 record: %s", v6Address.String())
							} else {
								log.Printf("[DynDNS Client] Successfully updated IPv6 record: %s", v6Address.String())
							}
						} else {
							log.Printf("[DynDNS Client] Failed to update IPv6 record: %s. Error: %s", v6Address.String(), result.V6.Error)
						}
					} else {
						log.Printf("[DynDNS Client] No IPv6 address provided. Skipping IPv6 update.")
					}

					return nil
				},
			},
		},
	}).Run(os.Args)

	if err != nil {
		log.Printf("[DynDNS Client] Error running client: %v", err)
	}

}
