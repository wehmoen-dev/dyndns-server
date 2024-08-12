package main

import (
	"dyndns/pkg/dns"
	"dyndns/pkg/server/routes"
	"dyndns/pkg/utils"
	"flag"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var bindAddress string
var auth string
var authFile string
var dnsZoneName string
var domainName string
var projectID string
var dynDNSService dns.DynDNSService

func init() {
	flag.StringVar(&bindAddress, "bind-address", utils.OsEnv("DYNDNS_BIND_ADDRESS", ":8080"), "Bind address for the server")
	flag.StringVar(&auth, "auth", os.Getenv("DYNDNS_AUTH"), "Basic Auth username:password")
	flag.StringVar(&authFile, "auth-file", utils.OsEnv("DYNDNS_AUTH_FILE", "google.json"), "Path to file containing the Google Cloud credentials")
	flag.StringVar(&dnsZoneName, "dns-zone-name", os.Getenv("DYNDNS_DNS_ZONE_NAME"), "DNS zone name")
	flag.StringVar(&domainName, "domain-name", os.Getenv("DYNDNS_DOMAIN_NAME"), "Domain name")
	flag.StringVar(&projectID, "project-id", os.Getenv("DYNDNS_PROJECT_ID"), "Google Cloud project ID")
	flag.Parse()

	if dnsZoneName == "" {
		log.Fatal("[DynDNS Server] DNS zone name is required")
	}

	if domainName == "" {
		log.Fatal("[DynDNS Server] Domain name is required")
	}

	if !strings.HasSuffix(domainName, ".") {
		domainName = domainName + "."
		log.Printf("[DynDNS Server] Appending '.' to domain name to get FQDN: %v", domainName)
	}

	if projectID == "" {
		log.Fatal("[DynDNS Server] Google Cloud project ID is required")
	}

	service, err := dns.NewService(authFile, projectID, dnsZoneName, domainName)

	if err != nil {
		log.Fatalf("[DynDNS Server] failed to create DNS service: %v", err)
	}

	dynDNSService = service
	err = dynDNSService.ValidateCredentials()

	if err != nil {
		log.Fatalf("[DynDNS Server] failed to validate credentials: %v", err)
	}

	log.Printf("[DynDNS Server] Credentials validated successfully!")

}

func main() {

	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.IPExtractor = echo.ExtractIPFromXFFHeader()
	server.Use(middleware.Recover())
	server.Use(middleware.RequestID())

	if auth != "" {
		server.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
			Skipper: func(c echo.Context) bool {
				return c.Path() == "/"
			},
			Validator: func(username, password string, c echo.Context) (bool, error) {
				if username+":"+password == auth {
					return true, nil
				}
				return false, nil
			},
		}))
	} else {
		log.Println("[DynDNS Server] Warning: No Basic Auth credentials. Waiting 5 seconds before starting server.")
		time.Sleep(5 * time.Second)
	}

	server.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Simple DynDNS Server for Google Cloud DNS")
	})

	routes.MountDynRoute(server, &routes.Config{
		DomainName: domainName,
		CloudDNS:   dynDNSService,
	})

	log.Printf("[DynDNS Server] Starting server on %v", bindAddress)

	err := server.Start(bindAddress)

	if err != nil {
		log.Fatalf("[DynDNS Server] failed to start server: %v", err)
	}

}
