package client

import (
	"github.com/go-resty/resty/v2"
	"net"
)

var IPGrabberOptions = map[string]*IPGrabberHosts{
	"icanhazipcom": {
		IPv4Host: "https://ipv4.icanhazip.com/",
		IPv6Host: "https://ipv6.icanhazip.com/",
	},
	"ipifyorg": {
		IPv4Host: "https://api.ipify.org/",
		IPv6Host: "https://api6.ipify.org/",
	},
	"ifconfigme": {
		IPv4Host: "https://ipv4.ifconfig.me/",
		IPv6Host: "https://ipv6.ifconfig.me/",
	},
	"ipsb": {
		IPv4Host: "https://api-ipv4.ip.sb/ip",
		IPv6Host: "https://api-ipv6.ip.sb/ip",
	},
	"identme": {
		IPv4Host: "https://v4.ident.me/",
		IPv6Host: "https://v6.ident.me/",
	},
}

type grabber struct {
	client *resty.Client
	hosts  *IPGrabberHosts
}

func (g *grabber) GrabV4() (*net.IP, error) {
	hosts := g.GetHosts()
	resp, err := g.client.R().Get(hosts.IPv4Host)

	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(resp.String())

	return &ip, err
}

func (g *grabber) GrabV6() (*net.IP, error) {
	hosts := g.GetHosts()

	resp, err := g.client.R().Get(hosts.IPv6Host)

	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(resp.String())

	return &ip, err
}

func (g *grabber) SetHosts(hosts *IPGrabberHosts) {
	g.hosts = hosts
}

func (g *grabber) GetHosts() *IPGrabberHosts {
	return g.hosts
}

func (g *grabber) GetAvailableHosts() []string {
	var hosts []string
	for k := range IPGrabberOptions {
		hosts = append(hosts, k)
	}
	return hosts
}

func NewIpGrabber(hosts *IPGrabberHosts) IpGrabber {

	client := resty.New()
	client.SetHeader("User-Agent", "dyndns-client/1.0 (https://github.com/wehmoen/dyndns-server)")
	client.SetDisableWarn(true)

	return &grabber{
		client: client,
		hosts:  hosts,
	}
}
