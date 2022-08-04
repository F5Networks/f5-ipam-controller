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

package provider

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/F5Networks/f5-ipam-controller/pkg/provider/sqlite"
	log "github.com/F5Networks/f5-ipam-controller/pkg/vlogger"
)

type IPAMProvider struct {
	store      StoreProvider
	ipamLabels map[string]bool
}

type Params struct {
	Range string
}

type StoreProvider interface {
	CreateTables() bool
	InsertIPs(ips []string, ipamLabel string)
	DisplayIPRecords()

	AllocateIP(ipamLabel, reference string) string
	ReleaseIP(ip string)
	GetIPAddressFromARecord(ipamLabel, hostname string) string
	GetIPAddressFromReference(ipamLabel, reference string) string

	CreateARecord(hostname, ipAddr string) bool
	DeleteARecord(hostname, ipAddr string) bool

	GetLabelMap() map[string]string
	AddLabel(label, ipRange string) bool
	RemoveLabel(label string) bool
	CleanUpLabel(label string)
}

func NewProvider(params Params) *IPAMProvider {
	// IPRangeMap := `{"test":"172.16.1.1-172.16.1.5", "prod":"172.16.1.50-172.16.1.55"}`

	prov := &IPAMProvider{
		store:      sqlite.NewStore(),
		ipamLabels: make(map[string]bool),
	}
	if !prov.Init(params) {
		log.Error("[PROV] Failed to Initialize Provider")
		return nil
	}
	log.Debugf("[PROV] Provider Initialised")
	return prov
}

func (prov *IPAMProvider) Init(params Params) bool {
	ipRangeMap := make(map[string]string)
	err := json.Unmarshal([]byte(params.Range), &ipRangeMap)
	if err != nil {
		log.Error("[PROV] Invalid IP range provided")
		return false
	}

	if prov.store == nil {
		log.Error("[PROV] Store not initialized")
		return false
	}

	labelMap := prov.store.GetLabelMap()

	for ipamLabel := range labelMap {
		if _, ok := ipRangeMap[ipamLabel]; !ok {
			// Remove all those labels from that are not present in the new ipRangeMap
			prov.store.CleanUpLabel(ipamLabel)
		}
	}

	for ipamLabel, ipRange := range ipRangeMap {

		// If the label exists in store, validate range and take corresponding action
		// if it doesn't exist in store, it is new label, create records by skipping "if" block
		if rng, ok := labelMap[ipamLabel]; ok {
			if rng == ipRange {
				// Exists and same range, nothing to do, simply skip to next
				prov.ipamLabels[ipamLabel] = true
				continue
			}
			// Exists and range changed, so remove range and add new range
			prov.store.CleanUpLabel(ipamLabel)
		}

		var ips []string
		for _, ipRangeItem := range strings.Split(ipRange, ",") {
			ipRangeConfig := strings.Split(ipRangeItem, "-")
			if len(ipRangeConfig) != 2 {
				log.Errorf("[PROV] Invalid IP range provided for %s label", ipamLabel)
				return false
			}

			startIP := net.ParseIP(ipRangeConfig[0])
			if startIP == nil {
				log.Errorf("[PROV] Invalid starting IP %s provided for %s label", ipRangeConfig[0], ipamLabel)
				return false
			}

			endIP := net.ParseIP(ipRangeConfig[1])
			if endIP == nil {
				log.Errorf("[PROV] Invalid ending IP %s provided for %s label", ipRangeConfig[1], ipamLabel)
				return false
			}

			for ; startIP.String() != endIP.String(); incIP(startIP) {
				ips = append(ips, startIP.String())
			}
			ips = append(ips, endIP.String())
		}

		prov.ipamLabels[ipamLabel] = true
		log.Debugf("Added Label: %v", ipamLabel)

		prov.store.AddLabel(ipamLabel, ipRange)
		prov.store.InsertIPs(ips, ipamLabel)
	}
	prov.store.DisplayIPRecords()

	return true
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
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

func (prov *IPAMProvider) GetIPAddressFromARecord(ipamLabel, hostname string) string {
	if _, ok := prov.ipamLabels[ipamLabel]; !ok {
		log.Debugf("[PROV] IPAM LABEL: %v Not Found", ipamLabel)
		return ""
	}
	return prov.store.GetIPAddressFromARecord(ipamLabel, hostname)
}

func (prov *IPAMProvider) GetIPAddressFromReference(ipamLabel, reference string) string {
	if _, ok := prov.ipamLabels[ipamLabel]; !ok {
		log.Debugf("[PROV] IPAM LABEL: %v Not Found", ipamLabel)
		return ""
	}
	return prov.store.GetIPAddressFromReference(ipamLabel, reference)
}

// Gets and reserves the next available IP address
func (prov *IPAMProvider) AllocateNextIPAddress(ipamLabel, reference string) string {
	if _, ok := prov.ipamLabels[ipamLabel]; !ok {
		log.Debugf("[PROV] Unsupported IPAM LABEL: %v", ipamLabel)
		return ""
	}
	return prov.store.AllocateIP(ipamLabel, reference)
}

// Releases an IP address
func (prov *IPAMProvider) ReleaseAddr(ipAddr string) {
	prov.store.ReleaseIP(ipAddr)
}
