package controller

import (
	"github.com/f5devcentral/f5-ipam-controller/pkg/ipamspec"
	"github.com/f5devcentral/f5-ipam-controller/pkg/manager"
	"github.com/f5devcentral/f5-ipam-controller/pkg/orchestration"
	log "github.com/f5devcentral/f5-ipam-controller/pkg/vlogger"
)

type Spec struct {
	Orchestrator orchestration.Orchestrator
	Manager      manager.Manager
	StopCh       chan struct{}
}

type Controller struct {
	Spec
	reqChan  chan ipamspec.IPAMRequest
	respChan chan ipamspec.IPAMResponse
}

func NewController(spec Spec) *Controller {
	ctlr := &Controller{
		Spec:     spec,
		reqChan:  make(chan ipamspec.IPAMRequest),
		respChan: make(chan ipamspec.IPAMResponse),
	}

	return ctlr
}

func (ctlr *Controller) runController() {
	for req := range ctlr.reqChan {
		switch req.Operation {
		case ipamspec.CREATE:

			sendResponse := func(request ipamspec.IPAMRequest, ipAddr string) {
				resp := ipamspec.IPAMResponse{
					Request: request,
					IPAddr:  ipAddr,
					Status:  true,
				}
				ctlr.respChan <- resp
			}

			// Controller tries to allocate asked IP Address to be allocated for the host from the give cidr
			// This happens during Starting of Controller to sync the DB with Initial Requests
			if req.IPAddr != "" {
				if ctlr.Manager.AllocateIPAddress(req.CIDR, req.IPAddr) {
					log.Debugf("[CORE] Allocated IP: %v for CIDR: %v", req.IPAddr, req.CIDR)
					ctlr.Manager.CreateARecord(req.HostName, req.IPAddr)
					go sendResponse(req, req.IPAddr)
				} else {
					log.Debugf("[CORE] Unable to Allocate asked IPAddress: %v to Host: %v in CIDR: %v",
						req.IPAddr, req.HostName, req.CIDR)
					go func(request ipamspec.IPAMRequest) {
						resp := ipamspec.IPAMResponse{
							Request: request,
							IPAddr:  "",
							Status:  false,
						}
						ctlr.respChan <- resp
					}(req)
				}
				break
			}

			ipAddr := ctlr.Manager.GetIPAddress(req.CIDR, req.HostName)
			if ipAddr != "" {
				go sendResponse(req, ipAddr)
				break
			}

			ipAddr = ctlr.Manager.GetNextIPAddress(req.CIDR)
			if ipAddr != "" {
				log.Debugf("[CORE] Allocated IP: %v for CIDR: %v", ipAddr, req.CIDR)
				ctlr.Manager.CreateARecord(req.HostName, ipAddr)
				go sendResponse(req, ipAddr)
			}
		case ipamspec.DELETE:
			ipAddr := ctlr.Manager.GetIPAddress(req.CIDR, req.HostName)
			if ipAddr != "" {
				ctlr.Manager.ReleaseIPAddress(ipAddr)
				ctlr.Manager.DeleteARecord(req.HostName, ipAddr)
			}
			go func(request ipamspec.IPAMRequest) {
				resp := ipamspec.IPAMResponse{
					Request: request,
					IPAddr:  "",
					Status:  true,
				}
				ctlr.respChan <- resp
			}(req)
		}
	}
}

func (ctlr *Controller) Start() {
	ctlr.Orchestrator.SetupCommunicationChannels(
		ctlr.reqChan,
		ctlr.respChan,
	)
	log.Info("[CORE] Controller started")

	ctlr.Orchestrator.Start(ctlr.StopCh)

	go ctlr.runController()
}

func (ctlr *Controller) Stop() {
	ctlr.Orchestrator.Stop()
}
