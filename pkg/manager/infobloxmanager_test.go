package manager

import (
	"errors"
	"github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"
	ibxclient "github.com/infobloxopen/infoblox-go-client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var DNSData = make(map[string]ibxclient.RecordA)
var HostData = make(map[string]string)
var IpList = []string{"10.0.0.1", "10.0.0.2"}
var index = 0

var _ = Describe("New infoblox manager ", func() {
	//request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: ""}
	It("Parsing JSON string provided in infoblox-label parameter ", func() {
		// Try with valid json in params
		labels, _ := ParseLabels(`{"Dev" :{"cidr": "172.16.4.0/24"},"Test" :{"cidr": "172.16.5.0/24"}}`)
		Expect(labels["Dev"]).To(BeEquivalentTo(IBConfig{"", "172.16.4.0/24"}))
		Expect(labels["Test"]).To(BeEquivalentTo(IBConfig{"", "172.16.5.0/24"}))
	})
	It("Infoblox manager client", func() {
		// Trying with invalid Json params
		infoParams := InfobloxParams{"localhost",
			"2.2.6",
			"6443",
			"admin",
			"infoblox",
			"{Dev :{\"cidr\": \"172.16.4.0/24\"},\"Test\" :{\"cidr\": \"172.16.5.0/24\"}}",
			"default",
			"false"}
		_, err := NewInfobloxManager(infoParams)
		Expect(err).NotTo(BeEquivalentTo(nil))
		// Try with valid json in params
		infoParams.IbLabelMap = "{\"Dev\" :{\"cidr\": \"172.16.4.0/24\"},\"Test\" :{\"cidr\": \"172.16.5.0/24\"}}"
		_, err = NewInfobloxManager(infoParams)
		Expect(err).NotTo(BeEquivalentTo(nil))
	})
})
var _ = Describe("Infoblox Manager functions", func() {
	infMgr := InfobloxManager{
		&ConnectorHandler{},
		&ObjMgrHandler{},
		ibxclient.EA{},
		"default",
		map[string]IBConfig{},
	}
	It("Testing CreateARecord function", func() {
		// Note: we are using the infMgr as defined in global section
		// let's start without any key and ipaddress in request it should fail
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: "Dev"}
		Expect(infMgr.CreateARecord(request)).To(BeFalse())
		// Now let's add a key and invalid ipaddress, it should fail this time as well
		request.IPAddr = "testing"
		request.HostName = "example.com"
		Expect(infMgr.CreateARecord(request)).To(BeFalse())
		// Now let us try with ipv6, it should fail this time as well
		request.IPAddr = "2000::ffff"
		Expect(infMgr.CreateARecord(request)).To(BeFalse())
		// Now let's add a valid ip address, this it should fail as iblabel map is not set
		request.IPAddr = "192.168.9.9"
		Expect(infMgr.CreateARecord(request)).To(BeFalse())
		// Now let's set the iblabel map
		infMgr.IBLabels["Dev"] = IBConfig{"", "192.168.9.0/24"}
		Expect(infMgr.CreateARecord(request)).To(BeTrue())
		Expect(len(DNSData)).To(BeEquivalentTo(1))
		// Now let's get the error from backend
		request.HostName = "send-error"
		Expect(infMgr.CreateARecord(request)).To(BeFalse())
	})
	It("Testing getARecords function", func() {
		// Note: we are using the infMgr as defined in global section
		// trying with invalid IPAM label
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: "invalid"}
		Expect(infMgr.getARecords(request)).To(BeEmpty())
		// Let's put the proper label in request
		request.IPAMLabel = "Dev"
		Expect(infMgr.getARecords(request)).To(BeEmpty())
		// Let's try to get the error if hostname does not exist
		request.HostName = "send-error"
		Expect(infMgr.getARecords(request)).To(BeEmpty())
		request.HostName = "example.com"
		result := infMgr.getARecords(request)
		Expect(result[0].Name).To(BeEquivalentTo("example.com"))
		Expect(result[0].Ipv4Addr).To(BeEquivalentTo("192.168.9.9"))
	})
	It("Testing DeleteARecord function", func() {
		// Note: we are using the infMgr as defined in global section
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.DELETE, HostName: "example.com", IPAddr: "192.168.9.9", Key: "", IPAMLabel: "Dev"}
		infMgr.DeleteARecord(request)
		Expect(len(DNSData)).To(BeEquivalentTo(0))
	})
	It("Testing validateIPAMLabels function", func() {
		// Note: we are using the infMgr as defined in global section
		// trying with valid CIDR
		result, _ := infMgr.validateIPAMLabels("", "192.168.9.9/24")
		Expect(result).To(BeTrue())
		result, _ = infMgr.validateIPAMLabels("", "send-error")
		Expect(result).To(BeFalse())
	})
	It("Testing AllocateNextIPAddress function", func() {
		// Note: we are using the infMgr as defined in global section
		// Let's try with invalid label first
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: "invalid"}
		Expect(infMgr.AllocateNextIPAddress(request)).To(BeEquivalentTo(""))
		Expect(len(HostData)).To(BeEquivalentTo(0))
		// Requesting error
		// Now let's set the iblabel map for sending the error
		infMgr.IBLabels["invalid"] = IBConfig{"", "send-error"}
		Expect(infMgr.AllocateNextIPAddress(request)).To(BeEquivalentTo(""))
		delete(infMgr.IBLabels, "invalid")
		// Now let's fix the label
		request.IPAMLabel = "Dev"
		// Let's add the hostname first to the request
		request.HostName = "foo.com"
		Expect(infMgr.AllocateNextIPAddress(request)).To(BeEquivalentTo(HostData[request.HostName]))
		Expect(len(HostData)).To(BeEquivalentTo(1))
		// Now let's try with key in request
		request.Key = "example-key"
		Expect(infMgr.AllocateNextIPAddress(request)).To(BeEquivalentTo(HostData[request.Key]))
		Expect(len(HostData)).To(BeEquivalentTo(2))
	})

	It("Testing GetIPAddress function", func() {
		// Note: we are using the infMgr as defined in global section
		// trying with invalid IPAM label
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: "invalid"}
		Expect(infMgr.GetIPAddress(request)).To(BeEquivalentTo(""))
		// Let's try with hostname first in request
		request.HostName = "foo.com"
		request.IPAMLabel = "Dev"
		Expect(infMgr.GetIPAddress(request)).To(BeEquivalentTo(HostData[request.HostName]))
		// Now let's try with key in request
		request.Key = "example-key"
		Expect(infMgr.GetIPAddress(request)).To(BeEquivalentTo(HostData[request.Key]))
	})

	It("Testing ReleaseIPAddress function", func() {
		// Note: we are using the infMgr as defined in global section
		// trying with invalid IPAM label
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: "invalid"}
		infMgr.ReleaseIPAddress(request)
		Expect(len(HostData)).To(BeEquivalentTo(2))
		// Let's try with hostname first in request
		request.HostName = "foo.com"
		request.IPAMLabel = "Dev"
		request.IPAddr = HostData[request.HostName]
		infMgr.ReleaseIPAddress(request)
		Expect(len(HostData)).To(BeEquivalentTo(1))
		// Requesting error
		request.IPAddr = "send-error"
		infMgr.ReleaseIPAddress(request)
		Expect(len(HostData)).To(BeEquivalentTo(1))
		// Now let's try with key in request
		request.Key = "example-key"
		request.IPAddr = HostData[request.Key]
		infMgr.ReleaseIPAddress(request)
		Expect(len(HostData)).To(BeEquivalentTo(0))

	})

})

func (manager ObjMgrHandler) GetNetwork(netview string, cidr string, ea ibxclient.EA) (*ibxclient.Network, error) {
	if cidr == "send-error" {
		return nil, errors.New("error as requested")
	}
	return &ibxclient.Network{}, nil
}

func (manager ObjMgrHandler) AllocateIP(netview string, cidr string, ipAddr string, macAddress string, name string, ea ibxclient.EA) (*ibxclient.FixedAddress, error) {
	if cidr == "send-error" {
		return nil, errors.New("error as requested")
	}
	HostData[name] = IpList[index]
	index += 1
	return &ibxclient.FixedAddress{NetviewName: netview, Cidr: cidr,
		IPAddress: HostData[name], Name: name, Ea: ea}, nil
}

func (manager ObjMgrHandler) ReleaseIP(netview string, cidr string, ipAddr string, macAddr string) (string, error) {
	if ipAddr == "send-error" {
		return "", errors.New("error as requested")
	}
	for k, v := range HostData {
		if v == ipAddr {
			delete(HostData, k)
		}
	}
	index -= 1
	return "", nil
}

func (manager ObjMgrHandler) DeleteARecord(ref string) (string, error) {
	delete(DNSData, ref)
	return ref, nil
}

func (manager ObjMgrHandler) CreateARecord(netview string, dnsview string, recordname string, cidr string, ipAddr string, ea ibxclient.EA) (*ibxclient.RecordA, error) {
	if recordname == "send-error" {
		return nil, errors.New("error as requested")
	}
	record := ibxclient.RecordA{Ref: recordname,
		Ipv4Addr: ipAddr,
		Name:     recordname,
		View:     netview,
		Zone:     dnsview,
		Ea:       ea,
	}
	DNSData[recordname] = record
	return &record, nil
}

func (connector ConnectorHandler) GetObject(obj ibxclient.IBObject, ref string, res interface{}) (err error) {
	switch obj.(type) {
	case *ibxclient.RecordA:
		rec := obj.(*ibxclient.RecordA)
		result := res.(*[]ibxclient.RecordA)
		if rec.Name == "send-error" {
			return errors.New("error as requested")
		}
		if rec.Name != "" {
			rec.Ref = rec.Name
			*result = append(*result, DNSData[rec.Name])
		}
	case *ibxclient.FixedAddress:
		rec := obj.(*ibxclient.FixedAddress)
		result := res.(*[]ibxclient.FixedAddress)
		for k, v := range HostData {
			tmpRec := rec
			(*tmpRec).IPAddress = v
			(*tmpRec).Name = k
			*result = append(*result, *tmpRec)
		}
	default:
		panic("Unexpected type")
	}
	return nil
}
