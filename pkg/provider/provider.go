package provider

import (
	"fmt"
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
	store *sqlite.DBStore
	cidrs map[string]bool
}

type Params struct {
	Range string
}

func NewProvider(params Params) *IPAMProvider {
	//ipArr := []string{"172.16.1.1-172.16.1.5/24", "172.16.1.50/22-172.16.1.55/22"}
	ipRanges := parseIPRange(params.Range)
	if ipRanges == nil {
		return nil
	}

	prov := &IPAMProvider{
		store: sqlite.NewStore(),
		cidrs: make(map[string]bool),
	}
	prov.generateExternalIPAddr(ipRanges)
	return prov

}

func parseIPRange(ipRange string) []string {
	if len(ipRange) == 0 {
		return nil
	}
	log.Debugf("[PROV] Parsing IP Ranges: %v", ipRange)
	ranges := strings.Split(ipRange, ",")
	var ipRanges []string
	for _, ipRange := range ranges {
		ipRanges = append(ipRanges, strings.Trim(ipRange, " "))
	}
	return ipRanges
}

// generateExternalIPAddr ...
func (prov *IPAMProvider) generateExternalIPAddr(ipRnages []string) {
	var startRangeIP, endRangeIP, Subnet string
	if len(ipRnages) == 0 {
		log.Fatal("[PROV] No IP range provided")
	}

	for _, ip := range ipRnages {
		ip = strings.Trim(ip, "\"")
		ipRangeArr := strings.Split(ip, "-")

		if len(ipRangeArr) != 2 {
			log.Errorf("Invalid IP Range Provided: %v", ip)
			continue
		}

		//checking the cidr of both the IPS if same then proceed otherwise error log
		ipRangeStart := strings.Split(ipRangeArr[0], "/")
		ipRangeEnd := strings.Split(ipRangeArr[1], "/")

		if len(ipRangeStart) != 2 || len(ipRangeEnd) != 2 {
			log.Errorf("Invalid IP Range Provided: %v", ip)
			continue
		}

		if ipRangeStart[1] != ipRangeEnd[1] {
			log.Debugf("[PROV] IPv4 Range Subnet mask is inconsistent")
			continue
		}
		switch ipv4or6(ip) {
		case IPV6:
			log.Debugf("[PROV] IPv6 is not supported")
		case IPV4:
			break
		default:
			log.Debugf("[PROV] Invalid IP Address provided in the range")
		}

		Subnet = ipRangeStart[1]

		startRangeIP = ipRangeStart[0]
		endRangeIP = ipRangeEnd[0]

		log.Debugf("[PROV] IP Pool: %v to %v/%v", startRangeIP, endRangeIP, Subnet)

		//endip validation
		ipEnd, ipNet, err := net.ParseCIDR(endRangeIP + "/" + Subnet)
		if err != nil {
			log.Debugf("[PROV] Parsing err :  ", err)
			continue
		}

		maskSize, _ := ipNet.Mask.Size()
		cidr := fmt.Sprintf("%s/%v", ipNet.IP.String(), maskSize)
		prov.cidrs[cidr] = true
		log.Debugf("[PROV] Processed CIDR: %v", cidr)

		//startip validation
		ipStart, ipnetStart, err := net.ParseCIDR(startRangeIP + "/" + Subnet)
		if err != nil {
			log.Debugf("[PROV] Parsing err : ", err)
			continue
		}
		ips := []string{}
		for ; ipnetStart.Contains(ipStart); inc(ipStart) {
			ips = append(ips, ipStart.String())
			// if len(ips) == EXTERNAL_IP_RANGE_COUNT {
			// 	break
			// }
			if ipStart.String() == ipEnd.String() {
				break
			}
		}
		prov.store.InsertIP(ips, cidr)
	}

	prov.store.DisplayIPRecords()
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

//external-ip-address parameter is of type ipv4 or ipv6
func ipv4or6(s string) string {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return IPV4
		case ':':
			return IPV6
		}
	}
	return "[PROV] Invalid Address"

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

func (prov *IPAMProvider) GetIPAddress(cidr, hostname string) string {
	if _, ok := prov.cidrs[cidr]; !ok {
		log.Debugf("[PROV] Unsupported CIDR: %v", cidr)
		return ""
	}
	ipAddr := prov.store.GetIPAddress(hostname)
	if doesCIDRContainIP(cidr, ipAddr) {
		return ipAddr
	}

	return ""
}

// Gets and reserves the next available IP address
func (prov *IPAMProvider) GetNextAddr(cidr string) string {
	if _, ok := prov.cidrs[cidr]; !ok {
		log.Debugf("[PROV] Unsupported CIDR: %v", cidr)
		return ""
	}
	return prov.store.AllocateIP(cidr)
}

// Marks an IP address as allocated if it belongs to that CIDR
func (prov *IPAMProvider) AllocateIPAddress(cidr, ipAddr string) bool {
	if _, ok := prov.cidrs[cidr]; !ok {
		log.Debugf("[PROV] Unsupported CIDR: %v", cidr)
		return false
	}

	if doesCIDRContainIP(cidr, ipAddr) {
		return prov.store.MarkIPAsAllocated(cidr, ipAddr)
	}
	return false
}

// Releases an IP address
func (prov *IPAMProvider) ReleaseAddr(ipAddr string) {
	prov.store.ReleaseIP(ipAddr)
}

func doesCIDRContainIP(cidr, ipAddr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Debugf("[PROV] Parsing CIDR error : ", err)
		return false
	}
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		log.Debugf("[PROV] Parsing IP error")
		return false
	}
	return ipNet.Contains(ip)
}
