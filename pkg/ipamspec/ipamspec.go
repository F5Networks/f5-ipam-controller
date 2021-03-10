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
	IPAMLabel string
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
		"\nHostname: %v\tKey: %v\tCIDR: %v\tIPAMLabel: %v\tIPAddr: %v\tOperation: %v\n",
		ipmReq.HostName,
		ipmReq.Key,
		ipmReq.CIDR,
		ipmReq.IPAMLabel,
		ipmReq.IPAddr,
		ipmReq.Operation,
	)
}
