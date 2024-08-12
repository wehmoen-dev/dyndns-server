package utils

import "net"

func IsReservedOrUnroutableIP(ip net.IP) bool {

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
