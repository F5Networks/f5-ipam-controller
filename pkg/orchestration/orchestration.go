package orchestration

import (
	"github.com/f5devcentral/f5-ipam-controller/pkg/ipamspec"
)

type Orchestrator interface {
	// SetupCommunicationChannels sets Request and Response channels
	SetupCommunicationChannels(reqChan chan<- ipamspec.IPAMRequest, respChan <-chan ipamspec.IPAMResponse)
	// Starts the Orchestrator, watching for resources
	Start(stopCh <-chan struct{})

	Stop()
}

func NewOrchestrator() Orchestrator {
	return NewIPAMK8SClient()
}
