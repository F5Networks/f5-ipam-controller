package ipamspec

import "fmt"

const (
	CREATE = "Create"
	DELETE = "Delete"
)

type IPAMRequest struct {
	Metadata  interface{}
	HostName  string
	Key       string
	CIDR      string
	IPAddr    string
	Operation string
}

type IPAMResponse struct {
	Request IPAMRequest
	IPAddr  string
	Status  bool
}

func (ipmReq IPAMRequest) String() string {
	return fmt.Sprintf(
		"\nHostname: %v\nKey: %v\nCIDR: %v\nIPAddr: %v\nOperation: %v\n",
		ipmReq.HostName,
		ipmReq.Key,
		ipmReq.CIDR,
		ipmReq.IPAddr,
		ipmReq.Operation,
	)
}
