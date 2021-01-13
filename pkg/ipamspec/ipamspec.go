package ipamspec

const (
	CREATE = "Create"
	DELETE = "Delete"
)

type IPAMRequest struct {
	Metadata  interface{}
	HostName  string
	CIDR      string
	IPAddr    string
	Operation string
}

type IPAMResponse struct {
	Request IPAMRequest
	IPAddr  string
	Status  bool
}
