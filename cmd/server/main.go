package main

import (
	"dyndns/pkg/dns"
	types "dyndns/pkg/server"
	"dyndns/pkg/utils"
	"errors"
	"flag"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net"
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

	server.GET("/dyn", func(c echo.Context) error {

		v4Address := c.QueryParam("ip_address")
		v6Address := c.QueryParam("ipv6_address")

		if v4Address == "" && v6Address == "" {
			log.Printf("[DynDNS Server][From:%s][Status:Error]: %s", c.RealIP(), "No IP address provided")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":  "No IP address provided",
				"detail": "Provide either an IPv4 (ip_address) or an IPv6 (ipv6_address) address or both to update the DNS record",
			})
		}

		result := types.UpdateResult{
			Name: domainName,
			V4: &dns.UpdateResult{
				RRType:  "A",
				Success: false,
				Created: false,
				Updated: false,
				Value:   "",
				Error:   nil,
			},
			V6: &dns.UpdateResult{
				RRType:  "AAAA",
				Success: false,
				Created: false,
				Updated: false,
				Value:   "",
				Error:   nil,
			},
		}

		if v4Address != "" {
			parsed := net.ParseIP(v4Address)

			if parsed == nil {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), v4Address, "Invalid IP address")
				result.V4.Error = errors.New("invalid IP address")
			}

			if parsed.IsLoopback() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Loopback IP address")
				result.V4.Error = errors.New("loopback IP address")
			}

			if parsed.IsUnspecified() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Unspecified IP address")
				result.V4.Error = errors.New("unspecified IP address")
			}

			if parsed.IsMulticast() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Multicast IP address")
				result.V4.Error = errors.New("multicast IP address")
			}

			if parsed.IsLinkLocalUnicast() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Link-local unicast IP address")
				result.V4.Error = errors.New("link-local unicast IP address")
			}

			if parsed.IsPrivate() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Private IP address")
				result.V4.Error = errors.New("private IP address")
			}

			if parsed.IsInterfaceLocalMulticast() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Interface-local multicast IP address")
				result.V4.Error = errors.New("interface-local multicast IP address")
			}

			if parsed.IsLinkLocalMulticast() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Link-local multicast IP address")
				result.V4.Error = errors.New("link-local multicast IP address")
			}

			if parsed.To4() == nil {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "IPv6 address")
				result.V4.Error = errors.New("IPv6 address")
			}

			if isReservedOrUnroutableIP(parsed) && result.V4.Error == nil {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Reserved or unroutable IP address")
				result.V4.Error = errors.New("reserved or unroutable IP address")
			}
		} else {
			result.V4.Error = errors.New("no IP address provided")
		}

		if v6Address != "" {
			parsed := net.ParseIP(v6Address)

			if parsed == nil {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), v6Address, "Invalid IP address")
				result.V6.Error = errors.New("invalid IP address")
			}

			if parsed.IsLoopback() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Loopback IP address")
				result.V6.Error = errors.New("loopback IP address")
			}

			if parsed.IsUnspecified() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Unspecified IP address")
				result.V6.Error = errors.New("unspecified IP address")
			}

			if parsed.IsMulticast() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Multicast IP address")
				result.V6.Error = errors.New("multicast IP address")
			}

			if parsed.IsLinkLocalUnicast() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Link-local unicast IP address")
				result.V6.Error = errors.New("link-local unicast IP address")
			}

			if parsed.IsPrivate() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Private IP address")
				result.V6.Error = errors.New("private IP address")
			}

			if parsed.IsInterfaceLocalMulticast() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Interface-local multicast IP address")
				result.V6.Error = errors.New("interface-local multicast IP address")
			}

			if parsed.IsLinkLocalMulticast() {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Link-local multicast IP address")
				result.V6.Error = errors.New("link-local multicast IP address")
			}

			if parsed.To16() == nil {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "IPv4 address")
				result.V6.Error = errors.New("IPv4 address")
			}

			if isReservedOrUnroutableIP(parsed) && result.V6.Error == nil {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Reserved or unroutable IP address")
				result.V6.Error = errors.New("reserved or unroutable IP address")
			}
		}

		if result.V4.Error != nil {
			v4Address = "" // Clear the IP address if it's invalid
		}

		if result.V6.Error != nil {
			v6Address = "" // Clear the IP address if it's invalid
		}

		v4Result, v6Result := dynDNSService.UpdateDNSRecord(v4Address, v6Address)

		if v4Result != nil {
			result.V4 = v4Result
			result.V4.Name = "" // Clear the domain name - its already in the parent struct
			if v4Result.Success {
				if v4Result.Created {
					log.Printf("[DynDNS Server][Type:A][From:%s][Status:Created][Domain:%s][IP:%s]: %s", c.RealIP(), domainName, v4Address, "DNS record created")
				} else if v4Result.Updated {
					log.Printf("[DynDNS Server][Type:A][From:%s][Status:Updated][Domain:%s][IP:%s]: %s", c.RealIP(), domainName, v4Address, "DNS record updated")
				}
			} else {
				if result.V4.Error == nil {
					result.V4.Error = errors.New("no IP address provided")
				}
			}
		}

		if v6Result != nil {
			result.V6 = v6Result
			result.V6.Name = "" // Clear the domain name - its already in the parent struct
			if v6Result.Success {
				if v6Result.Created {
					log.Printf("[DynDNS Server][Type:AAAA][From:%s][Status:Created][Domain:%s][IP:%s]: %s", c.RealIP(), domainName, v6Address, "DNS record created")
				} else if v6Result.Updated {
					log.Printf("[DynDNS Server][Type:AAAA][From:%s][Status:Updated][Domain:%s][IP:%s]: %s", c.RealIP(), domainName, v6Address, "DNS record updated")
				}
			} else {
				if result.V6.Error == nil {
					result.V6.Error = errors.New("no IP address provided")
				}
			}
		}

		return c.JSON(http.StatusOK, result)
	})

	log.Printf("[DynDNS Server] Starting server on %v", bindAddress)

	err := server.Start(bindAddress)

	if err != nil {
		log.Fatalf("[DynDNS Server] failed to start server: %v", err)
	}

}

func isReservedOrUnroutableIP(ip net.IP) bool {

	if ip.To4() == nil { // This checks if the IP is not IPv4 (i.e., it's IPv6)
		_, ipv6PublicRange, _ := net.ParseCIDR("2000::/3")
		return !ipv6PublicRange.Contains(ip)
	}

	reservedRanges := []string{
		"0.0.0.0/8",          // "This" Network
		"100.64.0.0/10",      // Shared Address Space
		"192.0.0.0/24",       // IETF Protocol Assignments
		"192.0.2.0/24",       // TEST-NET-1
		"198.18.0.0/15",      // Network Interconnect Device Benchmark Testing
		"198.51.100.0/24",    // TEST-NET-2
		"203.0.113.0/24",     // TEST-NET-3
		"240.0.0.0/4",        // Reserved for Future Use
		"255.255.255.255/32", // Limited Broadcast
	}

	for _, cidr := range reservedRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
