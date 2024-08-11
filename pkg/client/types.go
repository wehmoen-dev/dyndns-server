package client

import (
	"dyndns/pkg/server"
	"errors"
	"net"
)

var ArgNotFoundError = errors.New("arg not found")

type Arg struct {
	Name  string
	Value string
}

type Args []*Arg

func (a Args) Get(name string) (string, error) {
	for _, arg := range a {
		if arg.Name == name {
			return arg.Value, nil
		}
	}
	return "", ArgNotFoundError
}

type IPGrabberHosts struct {
	IPv4Host string
	IPv6Host string
}

type IpGrabber interface {
	GrabV4() (*net.IP, error)
	GrabV6() (*net.IP, error)
	SetHosts(hosts *IPGrabberHosts)
	GetHosts() *IPGrabberHosts
	GetAvailableHosts() []string
}

type RemoteApiCaller interface {
	Call(host string, v4Address *net.IP, v6Address *net.IP, auth ...string) (*server.UpdateResult, error)
}
