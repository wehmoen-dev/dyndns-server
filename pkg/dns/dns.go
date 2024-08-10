package dns

import (
	"context"
	"dyndns/pkg/utils"
	"fmt"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
	"time"
)

const (
	DNSCredentialValidationRecord = "_dyndns_credential_validation_record"
	DNSCredentialValidationIP     = "127.255.255.254"
)

var globalContext = context.Background()

type service struct {
	client      *dns.Service
	authFile    string
	projectID   string
	dnsZoneName string
	domainName  string
}

func NewService(authFile, projectID, dnsZoneName, domainName string) (DynDNSService, error) {

	credentials, err := utils.FindCredentials(authFile)

	if err != nil {
		return nil, err
	}

	dnsClient, err := dns.NewService(globalContext,
		option.WithScopes(dns.NdevClouddnsReadwriteScope),
		option.WithCredentialsJSON(credentials),
	)

	if err != nil {
		return nil, err
	}

	return &service{
		client:      dnsClient,
		authFile:    authFile,
		projectID:   projectID,
		dnsZoneName: dnsZoneName,
		domainName:  domainName,
	}, nil
}

func (s *service) UpdateDNSRecord(ipAddress string, ipv6Address string) (*UpdateResult, *UpdateResult) {

	if ipAddress == "" && ipv6Address == "" {
		return nil, nil
	}

	result := &UpdateResult{
		Name:    s.domainName,
		RRType:  "A",
		Success: false,
		Created: false,
		Updated: false,
		Value:   ipAddress,
		Error:   nil,
	}

	v6Result := &UpdateResult{
		Name:    s.domainName,
		RRType:  "AAAA",
		Success: false,
		Created: false,
		Updated: false,
		Value:   ipv6Address,
		Error:   nil,
	}

	if ipAddress != "" {
		if !s.recordExists(s.domainName, "A") {
			rrSet, err := s.client.ResourceRecordSets.Create(s.projectID, s.dnsZoneName, &dns.ResourceRecordSet{
				Kind:    "dns#resourceRecordSet",
				Name:    s.domainName,
				Type:    "A",
				Ttl:     1,
				Rrdatas: []string{ipAddress},
			}).Do()

			if err != nil {
				result.Error = fmt.Errorf("[DynDNS Server] failed to create resource record set: %v", err)
			}

			if rrSet.Rrdatas[0] != ipAddress {
				result.Error = fmt.Errorf("[DynDNS Server] failed to create resource record set: %v", err)
			}

			if err == nil && rrSet.Rrdatas[0] == ipAddress {
				result.Created = true
				result.Success = true
			}
		} else {
			rrSet, err := s.client.ResourceRecordSets.Patch(s.projectID, s.dnsZoneName, s.domainName, "A", &dns.ResourceRecordSet{
				Kind:    "dns#resourceRecordSet",
				Name:    s.domainName,
				Type:    "A",
				Ttl:     1,
				Rrdatas: []string{ipAddress},
			}).Do()

			if err != nil {
				result.Error = fmt.Errorf("[DynDNS Server] failed to patch resource record set: %v", err)
			}

			if rrSet.Rrdatas[0] != ipAddress {
				result.Error = fmt.Errorf("[DynDNS Server] failed to update resource record set: %v", err)
			}

			if err == nil && rrSet.Rrdatas[0] == ipAddress {
				result.Updated = true
				result.Success = true
			}
		}
	} else {
		result = nil
	}

	if ipv6Address != "" {
		if !s.recordExists(s.domainName, "AAAA") {
			rrSet, err := s.client.ResourceRecordSets.Create(s.projectID, s.dnsZoneName, &dns.ResourceRecordSet{
				Kind:    "dns#resourceRecordSet",
				Name:    s.domainName,
				Type:    "AAAA",
				Ttl:     1,
				Rrdatas: []string{ipv6Address},
			}).Do()

			if err != nil {
				v6Result.Error = fmt.Errorf("[DynDNS Server] failed to create resource record set: %v", err)
			}

			if rrSet.Rrdatas[0] != ipv6Address {
				v6Result.Error = fmt.Errorf("[DynDNS Server] failed to create resource record set: %v", err)
			}

			if err == nil && rrSet.Rrdatas[0] == ipv6Address {
				v6Result.Created = true
				v6Result.Success = true
			}
		} else {
			rrSet, err := s.client.ResourceRecordSets.Patch(s.projectID, s.dnsZoneName, s.domainName, "AAAA", &dns.ResourceRecordSet{
				Kind:    "dns#resourceRecordSet",
				Name:    s.domainName,
				Type:    "AAAA",
				Ttl:     1,
				Rrdatas: []string{ipv6Address},
			}).Do()

			if err != nil {
				v6Result.Error = fmt.Errorf("[DynDNS Server] failed to patch resource record set: %v", err)
			}

			if rrSet.Rrdatas[0] != ipv6Address {
				v6Result.Error = fmt.Errorf("[DynDNS Server] failed to update resource record set: %v", err)
			}

			if err == nil && rrSet.Rrdatas[0] == ipv6Address {
				v6Result.Updated = true
				v6Result.Success = true
			}
		}
	} else {
		v6Result = nil
	}

	return result, v6Result
}

func (s *service) ValidateCredentials() error {

	// Read Test
	_, err := s.client.ManagedZones.Get(s.projectID, s.dnsZoneName).Do()
	if err != nil {
		return fmt.Errorf("[DynDNS Server] failed to get managed zone: %v", err)
	}

	var testName = fmt.Sprintf("%s.%d.%s", DNSCredentialValidationRecord, time.Now().Unix(), s.domainName)

	// Write Test
	_, err = s.client.ResourceRecordSets.Create(s.projectID, s.dnsZoneName, &dns.ResourceRecordSet{
		Kind:    "dns#resourceRecordSet",
		Name:    testName,
		Type:    "A",
		Ttl:     300,
		Rrdatas: []string{DNSCredentialValidationIP},
	}).Do()

	if err != nil {
		return fmt.Errorf("[DynDNS Server] failed to patch resource record set: %v", err)
	}

	// Cleanup
	_, err = s.client.ResourceRecordSets.Delete(s.projectID, s.dnsZoneName, testName, "A").Do()

	if err != nil {
		return fmt.Errorf("[DynDNS Server] failed to delete resource record set: %v", err)
	}

	return nil
}

func (s *service) recordExists(name string, rrType string) bool {
	_, err := s.client.ResourceRecordSets.Get(s.projectID, s.dnsZoneName, name, rrType).Do()
	return err == nil
}

func (s *service) filterRecords(records []*dns.ResourceRecordSet, rrType string) []*dns.ResourceRecordSet {
	var filteredRecords []*dns.ResourceRecordSet
	for _, record := range records {
		if record.Type == rrType {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return filteredRecords
}
