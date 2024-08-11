package client

import (
	"dyndns/pkg/server"
	"errors"
	"github.com/go-resty/resty/v2"
	"log"
	"net"
	"net/http"
	"strings"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrInternalServerError = errors.New("internal server error")

type caller struct {
	client *resty.Client
}

func (c *caller) Call(host string, v4Address *net.IP, v6Address *net.IP, auth ...string) (*server.UpdateResult, error) {

	c.client.SetBaseURL(host)

	if len(auth) > 0 && len(auth[0]) > 0 {
		credentials := strings.SplitN(auth[0], ":", 2)
		if len(credentials) == 2 {
			c.client.SetBasicAuth(credentials[0], credentials[1])
		} else {
			log.Printf("[DynDNS Client] Could not parse credentials into username and password. Format should be username:password but received: %q", auth[0])
		}
	}

	var serverResponse server.UpdateResult

	var ipAddress string
	var ipv6Address string

	if v4Address != nil {
		ipAddress = v4Address.String()
	}

	if v6Address != nil {
		ipv6Address = v6Address.String()
	}

	response, err := c.client.R().
		SetResult(&serverResponse).
		SetQueryParams(map[string]string{
			"ip_address":   ipAddress,
			"ipv6_address": ipv6Address,
		}).Get("/dyn")

	if err != nil {
		return nil, err
	}

	if response.IsSuccess() {
		return &serverResponse, nil
	}

	if response.Error() != nil {
		return nil, response.Error().(error)
	}

	if response.StatusCode() == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if response.StatusCode() == http.StatusInternalServerError {
		return nil, ErrInternalServerError
	}

	return nil, nil // this is strange lol
}

func NewRemoteApiCaller() RemoteApiCaller {

	client := resty.New()
	client.SetDisableWarn(true)

	return &caller{
		client: client,
	}
}
