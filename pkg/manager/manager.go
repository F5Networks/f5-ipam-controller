package manager

import (
	"fmt"
	"github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
)

// Manager defines the interface that the IPAM system should implement
type Manager interface {
	// Creates an A record
	CreateARecord(req ipamspec.IPAMRequest) bool
	// Deletes an A record and releases the IP address
	DeleteARecord(req ipamspec.IPAMRequest)
	// Gets IP Address associated with hostname in the given CIDR
	GetIPAddress(req ipamspec.IPAMRequest) string
	// Gets and reserves the next available IP address in the given CIDR
	GetNextIPAddress(req ipamspec.IPAMRequest) string
	// Allocates given IP address from the CIDR
	AllocateIPAddress(req ipamspec.IPAMRequest) bool
	// Releases an IP address
	ReleaseIPAddress(req ipamspec.IPAMRequest)
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
