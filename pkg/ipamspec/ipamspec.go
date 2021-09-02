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

package ipamspec

import "fmt"

const (
	CREATE = "Create"
	DELETE = "Delete"
)

type IPAMRequest struct {
	Metadata  interface{}
	Operation string
	HostName  string
	IPAddr    string
	Key       string
	IPAMLabel string
}

type IPAMResponse struct {
	Request IPAMRequest
	IPAddr  string
	Status  bool
}

func (ipmReq IPAMRequest) String() string {
	return fmt.Sprintf(
		"\nHostname: %v\tKey: %v\tIPAMLabel: %v\tIPAddr: %v\tOperation: %v\n",
		ipmReq.HostName,
		ipmReq.Key,
		ipmReq.IPAMLabel,
		ipmReq.IPAddr,
		ipmReq.Operation,
	)
}
