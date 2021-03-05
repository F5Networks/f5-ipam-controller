package manager

import (
	"fmt"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
)

// Manager defines the interface that the IPAM system should implement
type Manager interface {
	// Creates an A record
	CreateARecord(hostname, ipAddr string) bool
	// Deletes an A record and releases the IP address
	DeleteARecord(hostname, ipAddr string)
	// Gets IP Address associated with hostname
	GetIPAddress(cidr, hostname string) string
	// Gets and reserves the next available IP address
	GetNextIPAddress(cidr string) string
	// Allocates this particular ip from the CIDR
	AllocateIPAddress(cidr, ipAddr string) bool
	// Releases an IP address
	ReleaseIPAddress(ipAddr string)
}

const F5IPAMProvider = "f5-ip-provider"

type Params struct {
	Provider string
	IPAMManagerParams
}

func NewManager(params Params) (Manager, error) {
	switch params.Provider {
	case F5IPAMProvider:
		log.Debugf("[MGR] Creating Manager with Provider: %v", F5IPAMProvider)
		f5IPAMParams := IPAMManagerParams{Range: params.Range}
		return NewIPAMManager(f5IPAMParams)
	default:
		log.Errorf("[MGR] Unknown Provider: %v", params.Provider)
	}
	return nil, fmt.Errorf("manager cannot be initialized")
}
