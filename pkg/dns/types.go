package dns

import "encoding/json"

type DynDNSService interface {
	UpdateDNSRecord(ipAddress string, ipv6Address string) (*UpdateResult, *UpdateResult)
	ValidateCredentials() error
}

type UpdateResult struct {
	Name    string `json:"name,omitempty"`
	RRType  string `json:"rr_type"`
	Success bool   `json:"success"`
	Created bool   `json:"created"`
	Updated bool   `json:"updated"`
	Value   string `json:"value"`
	Error   error  `json:"-"`
}

func (u UpdateResult) MarshalJSON() ([]byte, error) {
	type Alias UpdateResult
	return json.Marshal(&struct {
		*Alias
		Error string `json:"error,omitempty"`
	}{
		Error: func() string {
			if u.Error != nil {
				return u.Error.Error()
			}
			return ""
		}(),
		Alias: (*Alias)(&u),
	})
}
