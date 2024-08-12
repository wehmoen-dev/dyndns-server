package routes

import "dyndns/pkg/dns"

type Config struct {
	DomainName string
	CloudDNS   dns.DynDNSService
}
