package routes

import (
	"dyndns/pkg/dns"
	types "dyndns/pkg/server"
	"dyndns/pkg/utils"
	"errors"
	"github.com/labstack/echo/v4"
	"log"
	"net"
	"net/http"
)

func MountDynRoute(e *echo.Echo, cfg *Config) {
	e.GET("/dyn", func(c echo.Context) error {

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
			Name: cfg.DomainName,
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
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:-]: %s", c.RealIP(), "Invalid IP address")
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

			if utils.IsReservedOrUnroutableIP(parsed) && result.V4.Error == nil {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:%s]: %s", c.RealIP(), parsed.String(), "Reserved or unroutable IP address")
				result.V4.Error = errors.New("reserved or unroutable IP address")
			}
		} else {
			result.V4.Error = errors.New("no IP address provided")
		}

		if v6Address != "" {
			parsed := net.ParseIP(v6Address)

			if parsed == nil {
				log.Printf("[DynDNS Server][From:%s][Status:Error][IP:-]: %s", c.RealIP(), "Invalid IP address")
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

			if utils.IsReservedOrUnroutableIP(parsed) && result.V6.Error == nil {
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

		v4Result, v6Result := cfg.CloudDNS.UpdateDNSRecord(v4Address, v6Address)

		if v4Result != nil {
			result.V4 = v4Result
			result.V4.Name = "" // Clear the domain name - its already in the parent struct
			if v4Result.Success {
				parsed := net.ParseIP(v4Address)
				if v4Result.Created {
					log.Printf("[DynDNS Server][Type:A][From:%s][Status:Created][Domain:%s][IP:%s]: %s", c.RealIP(), cfg.DomainName, parsed.String(), "DNS record created")
				} else if v4Result.Updated {
					log.Printf("[DynDNS Server][Type:A][From:%s][Status:Updated][Domain:%s][IP:%s]: %s", c.RealIP(), cfg.DomainName, parsed.String(), "DNS record updated")
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
				parsed := net.ParseIP(v6Address)
				if v6Result.Created {
					log.Printf("[DynDNS Server][Type:AAAA][From:%s][Status:Created][Domain:%s][IP:%s]: %s", c.RealIP(), cfg.DomainName, parsed.String(), "DNS record created")
				} else if v6Result.Updated {
					log.Printf("[DynDNS Server][Type:AAAA][From:%s][Status:Updated][Domain:%s][IP:%s]: %s", c.RealIP(), cfg.DomainName, parsed.String(), "DNS record updated")
				}
			} else {
				if result.V6.Error == nil {
					result.V6.Error = errors.New("no IP address provided")
				}
			}
		}

		return c.JSON(http.StatusOK, result)
	})
}
