package server

import "dyndns/pkg/dns"

type UpdateResult struct {
	Name string            `json:"name"`
	V4   *dns.UpdateResult `json:"v4"`
	V6   *dns.UpdateResult `json:"v6"`
}
