/*-
 * Copyright (c) 2021, F5 Networks, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package manager

import (
	"fmt"
	"net"
	"strings"

	"github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"
	"github.com/F5Networks/f5-ipam-controller/pkg/provider"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
)

type IPAMManagerParams struct {
	Range string
}

type providerHandler struct {
	*provider.IPAMProvider
}

type IPAMManager struct {
	provider *providerHandler
}

func NewIPAMManager(params IPAMManagerParams) (*IPAMManager, error) {
	provParams := provider.Params{Range: params.Range}
	prov := provider.NewProvider(provParams)
	if prov == nil {
		return nil, fmt.Errorf("[IPMG] Unable to create Provider")
	}
	return &IPAMManager{provider: &providerHandler{prov}}, nil
}

// CreateARecord method creates an A record
func (ipMgr *IPAMManager) CreateARecord(req ipamspec.IPAMRequest) bool {
	if req.IPAddr == "" || (req.HostName == "" && req.Key == "") {
		log.Errorf("[IPMG] Invalid Request to Create A Record: %v", req.String())
		return false
	}
	if !isIPV4Addr(req.IPAddr) {
		log.Errorf("[IPMG] Unable to Create 'A' Record, as Invalid IP Address Provided")
		return false
	}
	if req.Key != "" {
		ipMgr.provider.CreateARecord(req.Key, req.IPAddr)
		return true
	}
	// TODO: Validate hostname to be a proper dns hostname
	ipMgr.provider.CreateARecord(req.HostName, req.IPAddr)
	return true
}

// DeleteARecord method deletes an A record and releases the IP address
func (ipMgr *IPAMManager) DeleteARecord(req ipamspec.IPAMRequest) {
	if req.IPAddr == "" || (req.HostName == "" && req.Key == "") {
		log.Errorf("[IPMG] Invalid Request to Delete A Record: %v", req.String())
	}
	if !isIPV4Addr(req.IPAddr) {
		log.Errorf("[IPMG] Unable to Delete 'A' Record, as Invalid IP Address Provided")
		return
	}
	if req.Key != "" {
		ipMgr.provider.DeleteARecord(req.Key, req.IPAddr)
		return
	}
	// TODO: Validate hostname to be a proper dns hostname
	ipMgr.provider.DeleteARecord(req.HostName, req.IPAddr)
}

func (ipMgr *IPAMManager) GetIPAddress(req ipamspec.IPAMRequest) string {
	if req.IPAMLabel == "" || (req.HostName == "" && req.Key == "") {
		log.Errorf("[IPMG] Invalid request to get IPAddress: %v", req.String())
		return ""
	}

	ref := req.HostName

	if ref == "" {
		ref = req.Key
	}

	return ipMgr.provider.GetIPAddressFromReference(req.IPAMLabel, ref)

}

// AllocateNextIPAddress method gets and reserves the next available IP address
func (ipMgr *IPAMManager) AllocateNextIPAddress(req ipamspec.IPAMRequest) string {
	ref := req.HostName

	if ref == "" {
		ref = req.Key
	}
	return ipMgr.provider.AllocateNextIPAddress(req.IPAMLabel, ref)
}

// ReleaseIPAddress method releases an IP address
func (ipMgr *IPAMManager) ReleaseIPAddress(req ipamspec.IPAMRequest) {
	if !isIPV4Addr(req.IPAddr) {
		log.Errorf("[IPMG] Unable to Release IP Address, as Invalid IP Address Provided")
		return
	}
	ipMgr.provider.ReleaseAddr(req.IPAddr)
}

func isIPV4Addr(ipAddr string) bool {
	if ipAddr == "" {
		return false
	}
	if net.ParseIP(ipAddr) == nil {
		return false
	}

	// presence of ":" indicates it is an IPV6
	if strings.Contains(ipAddr, ":") {
		return false
	}

	return true
}
