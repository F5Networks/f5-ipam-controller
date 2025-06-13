package manager

import (
	"github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var recordData = make(map[string]mockRecord)

var ipAddresses = []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"}
var ipindex = 0

type mockRecord struct {
	label     string
	reference string
	ipaddress string
}

var _ = Describe("Static IP Provider", func() {
	ipamMap := make(map[string]bool)
	ipamMap["dev"] = true
	ipamMap["test"] = true
	ipMgr := IPAMManager{provider: &providerHandler{}}

	It("Testing CreateARecord function", func() {
		// let's start without any key and ipaddress in request it should fail
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: ""}
		Expect(ipMgr.CreateARecord(request)).To(BeFalse())
		// Now let's add a key and invalid ipaddress, it should fail this time as well
		request.IPAddr = "testing"
		request.Key = "no-hostname"
		Expect(ipMgr.CreateARecord(request)).To(BeFalse())
		// Now let us try with ipv6, it should fail this time as well
		request.IPAddr = "2000::ffff"
		Expect(ipMgr.CreateARecord(request)).To(BeFalse())
		// Now let's add a valid ip address, this it should pass
		request.IPAddr = "192.168.9.9"
		//// create A record
		Expect(ipMgr.CreateARecord(request)).To(BeTrue())
		_, ok := recordData["no-hostname"]
		Expect(ok).To(BeTrue())
		// Now Let's create a record with hostname
		request.Key = ""
		request.HostName = "foo.com"
		request.IPAddr = "192.168.1.1"
		Expect(ipMgr.CreateARecord(request)).To(BeTrue())
		_, ok = recordData["foo.com"]
		Expect(ok).To(BeTrue())
	})
	It("Testing DeleteARecord function", func() {
		// let's start without any key and ipaddress in request it should fail
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: ""}
		ipMgr.DeleteARecord(request)
		Expect(len(recordData)).To(BeEquivalentTo(2))
		// Now let's add a key and invalid ipaddress, it should fail this time as well
		request.IPAddr = "testing"
		request.Key = "no-hostname"
		Expect(len(recordData)).To(BeEquivalentTo(2))
		// Now let's add a valid ip address, this it should pass
		request.IPAddr = "192.168.9.9"
		//// create A record
		ipMgr.DeleteARecord(request)
		Expect(len(recordData)).To(BeEquivalentTo(1))
		_, ok := recordData["no-hostname"]
		Expect(ok).To(BeFalse())
		// Now Let's delete a record with hostname
		request.Key = ""
		request.HostName = "foo.com"
		request.IPAddr = "192.168.1.1"
		ipMgr.DeleteARecord(request)
		Expect(len(recordData)).To(BeEquivalentTo(0))
		_, ok = recordData["foo.com"]
		Expect(ok).To(BeFalse())
	})
	It("Testing AllocateNextIPAddress function", func() {
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "foo.com", IPAddr: "", Key: "", IPAMLabel: "dev"}
		ipMgr.AllocateNextIPAddress(request)
		_, status := recordData["foo.com"]
		Expect(status).To(BeTrue())
		Expect(len(recordData)).To(BeEquivalentTo(1))
		request.Key = "no-hostname"
		request.HostName = ""
		request.IPAMLabel = "test"
		ipMgr.AllocateNextIPAddress(request)
		_, status = recordData["no-hostname"]
		Expect(status).To(BeTrue())
		Expect(len(recordData)).To(BeEquivalentTo(2))
	})
	It("Testing GetIPAddress function", func() {
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "", IPAMLabel: ""}
		ip := ipMgr.GetIPAddress(request)
		Expect(ip).To(BeEquivalentTo(""))
		request.IPAMLabel = "dev"
		request.Key = "no-hostname"
		// Get the ipaddress from valid label
		Expect(ipMgr.GetIPAddress(request)).To(Equal(recordData["no-hostname"].ipaddress))
		request.IPAMLabel = "test"
		request.HostName = "foo.com"
		request.Key = ""
		Expect(ipMgr.GetIPAddress(request)).To(Equal(recordData["foo.com"].ipaddress))
	})
	It("Testing ReleaseIPAddress function", func() {
		request := ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.DELETE, HostName: "", IPAddr: "test", Key: "", IPAMLabel: ""}
		ipMgr.ReleaseIPAddress(request)
		Expect(len(recordData)).To(BeEquivalentTo(2))
		request.IPAddr = recordData["no-hostname"].ipaddress
		ipMgr.ReleaseIPAddress(request)
		Expect(len(recordData)).To(BeEquivalentTo(1))
		request.IPAddr = recordData["foo.com"].ipaddress
		ipMgr.ReleaseIPAddress(request)
		Expect(len(recordData)).To(BeEquivalentTo(0))
	})

})

func (manager providerHandler) CreateARecord(hostname, ipAddr string) bool {
	recordData[hostname] = mockRecord{"", hostname, ipAddr}
	return true
}

func (manager providerHandler) DeleteARecord(hostname, ipAddr string) {
	delete(recordData, hostname)
}

func (manager providerHandler) GetIPAddressFromARecord(ipamLabel, hostname string) string {
	return recordData[hostname].ipaddress
}

func (manager providerHandler) GetIPAddressFromReference(ipamLabel, reference string) string {
	return recordData[reference].ipaddress
}

func (manager providerHandler) AllocateNextIPAddress(ipamLabel, reference string) string {
	recordData[reference] = mockRecord{ipamLabel,
		reference,
		ipAddresses[ipindex],
	}
	ipindex += 1
	return ipAddresses[ipindex]
}

func (manager providerHandler) ReleaseAddr(ipAddr string) {
	for k, v := range recordData {
		if v.ipaddress == ipAddr {
			delete(recordData, k)
		}
	}
	ipindex -= 1
}
