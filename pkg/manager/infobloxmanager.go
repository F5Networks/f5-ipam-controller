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
	NetView    string
	SslVerify  string
}

type ObjMgrHandler struct {
	*ibxclient.ObjectManager
}

type ConnectorHandler struct {
	*ibxclient.Connector
}

type InfobloxManager struct {
	connector *ConnectorHandler
	objMgr    *ObjMgrHandler
	ea        ibxclient.EA
	NetView   string
	IBLabels  map[string]IBConfig
}

type IBConfig struct {
	DNSView string `json:"dnsView,omitempty"`
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

	labels, err := ParseLabels(params.IbLabelMap)
	if err != nil {
		return nil, err
	}

	// TransportConfig params: sslVerify, httpRequestsTimeout, httpPoolConnections
	// These are the common values
	transportConfig := ibxclient.NewTransportConfig(params.SslVerify, 20, 10)
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

	ibMgr := &InfobloxManager{
		connector: &ConnectorHandler{connector},
		objMgr:    &ObjMgrHandler{objMgr},
		ea:        ibxclient.EA{EAKey: EAVal},
		IBLabels:  labels,
		NetView:   params.NetView,
	}
	_, err = ibMgr.objMgr.GetNetworkView(ibMgr.NetView)
	if err != nil {
		return nil, err
	}
	// Validating that dnsView, CIDR exist on infoblox Server
	for _, parameter := range labels {
		result, err := ibMgr.validateIPAMLabels(parameter.DNSView, parameter.CIDR)
		if !result {
			return nil, err
		}
	}
	return ibMgr, nil
}

func ParseLabels(params string) (map[string]IBConfig, error) {
	ibLabelMap := make(map[string]IBConfig)
	err := json.Unmarshal([]byte(params), &ibLabelMap)
	if err != nil {
		return nil, err
	}
	for label, ibParam := range ibLabelMap {
		// DNSView is being disabled
		// The below line can be removed when DNSView support is enabled
		ibParam.DNSView = ""

		ibLabelMap[label] = ibParam
	}
	return ibLabelMap, nil
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

	label, ok := infMgr.IBLabels[req.IPAMLabel]
	if !ok {
		return false
	}

	_, err := infMgr.objMgr.CreateARecord(
		infMgr.NetView,
		label.DNSView,
		req.HostName,
		label.CIDR,
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
	res := infMgr.getARecords(req)

	_, err := infMgr.objMgr.DeleteARecord(res[0].Ref)
	if err != nil {
		log.Errorf("[IPMG] 'A' Record not available, %+v", req)
	}
}

// GetIPAddress Gets IP Address associated with hostname
func (infMgr *InfobloxManager) GetIPAddress(req ipamspec.IPAMRequest) string {
	if req.HostName == "" && req.Key == "" {
		log.Errorf("[IPMG] Invalid Request to get IPAddress: %+v", req)
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

	ip := infMgr.getIPAddressFromName(req)

	return ip
}

// GetNextIPAddress Gets and reserves the next available IP address
func (infMgr *InfobloxManager) AllocateNextIPAddress(req ipamspec.IPAMRequest) string {
	label, ok := infMgr.IBLabels[req.IPAMLabel]
	if !ok {
		return ""
	}
	name := req.HostName
	if req.Key != "" {
		name = req.Key
	}
	fixedAddr, err := infMgr.objMgr.AllocateIP(infMgr.NetView, label.CIDR, "", "", name, infMgr.ea)
	if err != nil {
		log.Errorf("[IPMG] Unable to Get a New IP Address: %+v", req)
		return ""
	}
	return fixedAddr.IPAddress
}

// ReleaseIPAddress Releases an IP address
func (infMgr *InfobloxManager) ReleaseIPAddress(req ipamspec.IPAMRequest) {
	label, ok := infMgr.IBLabels[req.IPAMLabel]
	if !ok {
		return
	}
	_, err := infMgr.objMgr.ReleaseIP(infMgr.NetView, label.CIDR, req.IPAddr, "")
	if err != nil {
		log.Errorf("[IPMG] Unable to Release IP Address: %+v", req)
	}
	return
}

func (infMgr *InfobloxManager) getARecords(req ipamspec.IPAMRequest) []ibxclient.RecordA {
	var res []ibxclient.RecordA

	label, ok := infMgr.IBLabels[req.IPAMLabel]
	if !ok {
		return nil
	}

	recA := ibxclient.NewRecordA(ibxclient.RecordA{
		Name: req.HostName,
		View: label.DNSView,
	})

	err := infMgr.connector.GetObject(recA, "", &res)
	if err != nil || len(res) == 0 {
		log.Errorf("[IPMG] 'A' Record not available, %+v", req)
		return nil
	}
	return res
}

func (infMgr *InfobloxManager) getIPAddressFromName(req ipamspec.IPAMRequest) (ip string) {
	var returnFixedAddresses []ibxclient.FixedAddress

	label, ok := infMgr.IBLabels[req.IPAMLabel]
	if !ok {
		return ""
	}

	name := req.HostName
	if req.Key != "" {
		name = req.Key
	}

	fixedAddr := ibxclient.NewFixedAddress(ibxclient.FixedAddress{
		NetviewName: infMgr.NetView,
		Cidr:        label.CIDR,
	})

	err := infMgr.connector.GetObject(fixedAddr, "", &returnFixedAddresses)

	if err != nil || returnFixedAddresses == nil || len(returnFixedAddresses) == 0 {
		log.Errorf("[Infoblox] IP not available, %+v", req)
		return ""
	}

	for _, fixedAddress := range returnFixedAddresses {
		if fixedAddress.Name == name {
			return fixedAddress.IPAddress
		}
	}
	return ""
}

func (infMgr *InfobloxManager) validateIPAMLabels(dnsView, cidr string) (bool, error) {

	// DNSView is being disabled
	// The below code can be uncommented when DNSView support is enabled
	//if len(dnsView) == 0 {
	//	return false, fmt.Errorf("dnsView should not be empty")
	//}
	_, err := infMgr.objMgr.GetNetwork(infMgr.NetView, cidr, infMgr.ea)
	if err != nil {
		return false, err
	}
	return true, nil
}
