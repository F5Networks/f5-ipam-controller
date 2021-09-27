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

package orchestration

import (
	"github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"
)

type Orchestrator interface {
	// SetupCommunicationChannels sets Request and Response channels
	SetupCommunicationChannels(reqChan chan<- ipamspec.IPAMRequest, respChan <-chan ipamspec.IPAMResponse)

	// Start starts the Orchestrator, watching for resources
	Start(stopCh <-chan struct{})

	Stop()
}

func NewOrchestrator() Orchestrator {
	return NewIPAMK8SClient()
}
