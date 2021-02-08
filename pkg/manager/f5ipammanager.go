package manager

import (
	"fmt"
	"net"
	"strings"

	"github.com/f5devcentral/f5-ipam-controller/pkg/provider"
	log "github.com/f5devcentral/f5-ipam-controller/pkg/vlogger"
)

type IPAMManagerParams struct {
	Range string
}

type IPAMManager struct {
	provider *provider.IPAMProvider
}

func NewIPAMManager(params IPAMManagerParams) (*IPAMManager, error) {
	provParams := provider.Params{Range: params.Range}
	prov := provider.NewProvider(provParams)
	if prov == nil {
		return nil, fmt.Errorf("[IPMG] Unable to create Provider")
	}
	return &IPAMManager{provider: prov}, nil
}

// Creates an A record
func (ipMgr *IPAMManager) CreateARecord(hostname, ipAddr string) bool {
	if !isIPV4Addr(ipAddr) {
		log.Errorf("[IPMG] Invalid IP Address Provided")
		return false
	}
	// TODO: Validate hostname to be a proper dns hostname
	ipMgr.provider.CreateARecord(hostname, ipAddr)
	return true
}

// Deletes an A record and releases the IP address
func (ipMgr *IPAMManager) DeleteARecord(hostname, ipAddr string) {
	if !isIPV4Addr(ipAddr) {
		log.Errorf("[IPMG] Invalid IP Address Provided")
		return
	}
	// TODO: Validate hostname to be a proper dns hostname
	ipMgr.provider.DeleteARecord(hostname, ipAddr)
}

func (ipMgr *IPAMManager) GetIPAddress(hostname string) string {
	// TODO: Validate hostname to be a proper dns hostname
	return ipMgr.provider.GetIPAddress(hostname)
}

// Gets and reserves the next available IP address
func (ipMgr *IPAMManager) GetNextIPAddress(cidr string) string {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Debugf("[IPMG] Invalid CIDR Provided: %v", cidr)
		return ""
	}
	return ipMgr.provider.GetNextAddr(cidr)
}

// Allocates this particular ip from the CIDR
func (ipMgr *IPAMManager) AllocateIPAddress(cidr, ipAddr string) bool {
	return ipMgr.provider.AllocateIPAddress(cidr, ipAddr)
}

// Releases an IP address
func (ipMgr *IPAMManager) ReleaseIPAddress(ipAddr string) {

	if !isIPV4Addr(ipAddr) {
		log.Errorf("[IPMG] Invalid IP Address Provided")
		return
	}
	ipMgr.provider.ReleaseAddr(ipAddr)
}

func isIPV4Addr(ipAddr string) bool {
	if net.ParseIP(ipAddr) == nil {
		return false
	}

	// presence of ":" indicates it is an IPV6
	if strings.Contains(ipAddr, ":") {
		return false
	}

	return true
}
