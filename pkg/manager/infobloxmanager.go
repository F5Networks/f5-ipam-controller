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
	"encoding/json"
	"fmt"

	"github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"
	"github.com/F5Networks/f5-ipam-controller/pkg/utils"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
	ibxclient "github.com/infobloxopen/infoblox-go-client"
)

const (
	EAKey = "F5IPAM"
	EAVal = "managed"
)

type InfobloxParams struct {
	Host       string
	Version    string
	Port       string
	Username   string
	Password   string
	IbLabelMap string
}

type InfobloxManager struct {
	connector *ibxclient.Connector
	objMgr    *ibxclient.ObjectManager
	ea        ibxclient.EA
	IBLabels  map[string]IBParam
}

type IBParam struct {
	NetView string `json:"netView"`
	DNSView string `json:"dnsView"`
	CIDR    string `json:"cidr"`
}

func NewInfobloxManager(params InfobloxParams) (*InfobloxManager, error) {
	hostConfig := ibxclient.HostConfig{
		Host:     params.Host,
		Version:  params.Version,
		Port:     params.Port,
		Username: params.Username,
		Password: params.Password,
	}

	// TransportConfig params: sslVerify, httpRequestsTimeout, httpPoolConnections
	// These are the common values
	transportConfig := ibxclient.NewTransportConfig("false", 20, 10)
	requestBuilder := &ibxclient.WapiRequestBuilder{}
	requestor := &ibxclient.WapiHttpRequestor{}
	connector, err := ibxclient.NewConnector(hostConfig, transportConfig, requestBuilder, requestor)
	if err != nil {
		return nil, err
	}
	objMgr := ibxclient.NewObjectManager(connector, "F5IPAM", "0")

	objMgr.OmitCloudAttrs = true

	// Create an Extensible Attribute for resource tracking
	if eaDef, _ := objMgr.GetEADefinition(EAKey); eaDef == nil {
		eaDef := ibxclient.EADefinition{
			Name:    EAKey,
			Type:    "STRING",
			Comment: "Managed by the F5 IPAM Controller",
		}
		_, err = objMgr.CreateEADefinition(eaDef)
		if err != nil {
			return nil, err
		}
	}

	labels, err := ParseLabels(params.IbLabelMap)
	if err != nil {
		return nil, err
	}

	ibMgr := &InfobloxManager{
		connector: connector,
		objMgr:    objMgr,
		ea:        ibxclient.EA{EAKey: EAVal},
		IBLabels:  labels,
	}

	// Validating that netView, dnsView, CIDR exist on infoblox Server
	for _, parameter := range labels {
		result, err := ibMgr.validateIPAMLabels(parameter.NetView, parameter.DNSView, parameter.CIDR)
		if !result {
			return nil, err
		}
	}
	return ibMgr, nil
}

func ParseLabels(params string) (map[string]IBParam, error) {
	ibLabelMap := make(map[string]IBParam)
	err := json.Unmarshal([]byte(params), &ibLabelMap)
	if err != nil {
		return nil, err
	}
	for label, ibParam := range ibLabelMap {
		ibLabelMap[label] = ibParam
	}
	return ibLabelMap, nil
}

func (infMgr *InfobloxManager) IsPersistent() bool {
	return true
}

// CreateARecord Creates an A record
func (infMgr *InfobloxManager) CreateARecord(req ipamspec.IPAMRequest) bool {
	if req.IPAddr == "" || req.HostName == "" {
		log.Errorf("[IPMG] Invalid Request to Create A Record: %v", req.String())
		return false
	}
	if !utils.IsIPAddr(req.IPAddr) {
		log.Errorf("[IPMG] Unable to Create 'A' Record, as Invalid IP Address Provided")
		return false
	}

	if ok := infMgr.getIBParams(&req); !ok {
		return false
	}

	_, err := infMgr.objMgr.CreateARecord(
		req.NetView,
		req.DNSView,
		req.HostName,
		req.CIDR,
		req.IPAddr,
		infMgr.ea,
	)
	if err != nil {
		log.Errorf("[IPMG] Unable to Create 'A' Record. Error: %v", err)
		return false
	}

	return true
}

// DeleteARecord Deletes an A record and releases the IP address
func (infMgr *InfobloxManager) DeleteARecord(req ipamspec.IPAMRequest) {
	if ok := infMgr.getIBParams(&req); !ok {
		return
	}
	res := infMgr.getARecords(req)

	_, err := infMgr.objMgr.DeleteARecord(res[0].Ref)
	if err != nil {
		log.Errorf("[IPMG] 'A' Record not available, %+v", req)
	}
}

// GetIPAddress Gets IP Address associated with hostname
func (infMgr *InfobloxManager) GetIPAddress(req ipamspec.IPAMRequest) string {
	if req.HostName == "" {
		log.Errorf("[IPMG] Invalid Request to Get IP Address: %+v", req)
		return ""
	}

	//hostRecord, err := infMgr.objMgr.GetHostRecord(req.HostName)
	//if err != nil {
	//	log.Errorf("[IPMG] No A Record available with Hostname to Get IP Address: %v", req.String())
	//	return ""
	//}
	//
	//ipAddr, err := infMgr.objMgr.GetIpAddressFromHostRecord(*hostRecord)
	//if err != nil {
	//	log.Errorf("[IPMG] No IP address available with Hostname to Get IP Address: %v", req.String())
	//	return ""
	//}
	if ok := infMgr.getIBParams(&req); !ok {
		return ""
	}
	res := infMgr.getARecords(req)

	if len(res) == 0 {
		return ""
	}

	return res[0].Ipv4Addr
}

// GetNextIPAddress Gets and reserves the next available IP address
func (infMgr *InfobloxManager) GetNextIPAddress(req ipamspec.IPAMRequest) string {
	if ok := infMgr.getIBParams(&req); !ok {
		return ""
	}
	fixedAddr, err := infMgr.objMgr.AllocateIP(req.NetView, req.CIDR, "", "", "", infMgr.ea)
	if err != nil {
		log.Errorf("[IPMG] Unable to Get a New IP Address: %+v", req)
		return ""
	}
	return fixedAddr.IPAddress
}

// AllocateIPAddress Allocates given IP address
func (infMgr *InfobloxManager) AllocateIPAddress(req ipamspec.IPAMRequest) bool {
	//_, err := infMgr.objMgr.AllocateIP(req.NetView, req.CIDR, req.IPAddr, "", "", infMgr.ea)
	//if err != nil {
	//	log.Errorf("[IPMG] Unable to Get a New IP Address: %v", req.String())
	//	return false
	//}
	return true
}

// ReleaseIPAddress Releases an IP address
func (infMgr *InfobloxManager) ReleaseIPAddress(req ipamspec.IPAMRequest) {
	if ok := infMgr.getIBParams(&req); !ok {
		return
	}
	_, err := infMgr.objMgr.ReleaseIP(req.NetView, req.CIDR, req.IPAddr, "")
	if err != nil {
		log.Errorf("[IPMG] Unable to Release IP Address: %+v", req)
	}
	return
}

func (infMgr *InfobloxManager) getARecords(req ipamspec.IPAMRequest) []ibxclient.RecordA {
	var res []ibxclient.RecordA

	if ok := infMgr.getIBParams(&req); !ok {
		return nil
	}

	recA := ibxclient.NewRecordA(ibxclient.RecordA{
		Name: req.HostName,
		View: req.DNSView,
	})

	err := infMgr.connector.GetObject(recA, "", &res)
	if err != nil || len(res) == 0 {
		log.Errorf("[IPMG] 'A' Record not available, %+v", req)
		return nil
	}
	return res
}

func (infMgr *InfobloxManager) validateIPAMLabels(netView, dnsView, cidr string) (bool, error) {
	_, err := infMgr.objMgr.GetNetworkView(netView)
	if err != nil {
		return false, err
	}
	if len(dnsView) == 0 {
		return false, fmt.Errorf("dnsView should not be empty")
	}
	_, err = infMgr.objMgr.GetNetwork(netView, cidr, infMgr.ea)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (infMgr *InfobloxManager) getIBParams(req *ipamspec.IPAMRequest) bool {
	label, ok := infMgr.IBLabels[req.IPAMLabel]
	if !ok {
		log.Errorf("[IPMG] Invalid Label: %v provided.", req.IPAMLabel)
		return false
	}
	req.NetView = label.NetView
	req.DNSView = label.DNSView
	req.CIDR = label.CIDR
	return true
}
