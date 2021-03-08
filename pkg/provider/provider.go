package provider

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/F5Networks/f5-ipam-controller/pkg/provider/sqlite"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
)

const (
	IPV4 = "IPv4"
	IPV6 = "IPv6"
)

type IPAMProvider struct {
	store    *sqlite.DBStore
	cidrTags map[string]bool
}

type Params struct {
	Range string
}

func NewProvider(params Params) *IPAMProvider {
	// IPRangeMap := `{"tag1":"172.16.1.1-172.16.1.5/24", "tag2":"172.16.1.50/22-172.16.1.55/22"}`

	prov := &IPAMProvider{
		store:    sqlite.NewStore(),
		cidrTags: make(map[string]bool),
	}
	if !prov.Init(params) {
		return nil
	}
	return prov
}

func (prov *IPAMProvider) Init(params Params) bool {
	ipRangeMap := make(map[string]string)
	err := json.Unmarshal([]byte(params.Range), &ipRangeMap)
	if err != nil {
		log.Fatal("[PROV] Invalid IP range provided")
		return false
	}
	for cidrTag, ipRange := range ipRangeMap {
		ipRangeConfig := parseIPRange(ipRange)
		if ipRangeConfig == nil {
			return false
		}

		startIP, netw, err := net.ParseCIDR(ipRangeConfig[0] + "/" + ipRangeConfig[2])
		if err != nil {
			return false
		}

		endIP, netw, err := net.ParseCIDR(ipRangeConfig[1] + "/" + ipRangeConfig[2])
		if err != nil {
			return false
		}

		var ips []string
		for ; netw.Contains(startIP); incIP(startIP) {

			ips = append(ips, startIP.String())
			if startIP.String() == endIP.String() {
				break
			}
		}

		prov.store.InsertIP(ips, cidrTag)
	}
	prov.store.DisplayIPRecords()

	return true
}

func parseIPRange(ipRange string) []string {
	if len(ipRange) == 0 {
		return nil
	}
	rangeBoundaries := strings.Split(ipRange, "-")
	if len(rangeBoundaries) != 2 {
		return nil
	}

	var startIP, endIP string

	ipConfig := strings.Split(rangeBoundaries[0], "/")
	switch isIPv4orv6(ipConfig[0]) {
	case IPV4:
		startIP = ipConfig[0]
	default:
		return nil
	}
	_, _, err := net.ParseCIDR(startIP + "/" + ipConfig[1])
	if err != nil {
		return nil
	}

	ipConfig = strings.Split(rangeBoundaries[1], "/")
	switch isIPv4orv6(ipConfig[0]) {
	case IPV4:
		endIP = ipConfig[0]
	default:
		return nil
	}

	_, _, err = net.ParseCIDR(endIP + "/" + ipConfig[1])
	if err != nil {
		return nil
	}

	return []string{startIP, endIP, ipConfig[1]}

}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

//external-ip-address parameter is of type ipv4 or ipv6
func isIPv4orv6(s string) string {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return IPV4
		case ':':
			return IPV6
		}
	}
	return ""

}

// Creates an A record
func (prov *IPAMProvider) CreateARecord(hostname, ipAddr string) bool {
	prov.store.CreateARecord(hostname, ipAddr)
	log.Debugf("[PROV] Created 'A' Record. Host:%v, IP:%v", hostname, ipAddr)
	return true
}

// Deletes an A record and releases the IP address
func (prov *IPAMProvider) DeleteARecord(hostname, ipAddr string) {
	prov.store.DeleteARecord(hostname, ipAddr)
	log.Debugf("[PROV] Deleted 'A' Record. Host:%v, IP:%v", hostname, ipAddr)
}

func (prov *IPAMProvider) GetIPAddress(cidrTag, hostname string) string {
	if _, ok := prov.cidrTags[cidrTag]; !ok {
		log.Debugf("[PROV] CIDR TAG: %v Not Found", cidrTag)
		return ""
	}
	return prov.store.GetIPAddress(cidrTag, hostname)
}

// Gets and reserves the next available IP address
func (prov *IPAMProvider) GetNextAddr(cidrTag string) string {
	if _, ok := prov.cidrTags[cidrTag]; !ok {
		log.Debugf("[PROV] Unsupported CIDR TAG: %v", cidrTag)
		return ""
	}
	return prov.store.AllocateIP(cidrTag)
}

// Marks an IP address as allocated if it belongs to that CIDR
func (prov *IPAMProvider) AllocateIPAddress(cidrTag, ipAddr string) bool {
	if _, ok := prov.cidrTags[cidrTag]; !ok {
		log.Debugf("[PROV] Unsupported CIDR TAG: %v", cidrTag)
		return false
	}

	return prov.store.MarkIPAsAllocated(cidrTag, ipAddr)
}

// Releases an IP address
func (prov *IPAMProvider) ReleaseAddr(ipAddr string) {
	prov.store.ReleaseIP(ipAddr)
}
