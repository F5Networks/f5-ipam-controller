package provider

import (
	"github.com/F5Networks/f5-ipam-controller/pkg/provider/sqlite/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Contoller Suite")
}

func ipRangeHelper(iprange string, result bool) (prov *IPAMProvider) {
	params := Params{
		Range: iprange,
	}
	prov = &IPAMProvider{
		store:      mock.NewMockStore(mock.MockData{}),
		ipamLabels: make(map[string]bool),
	}
	if result {
		Expect(prov.Init(params)).To(BeTrue())
	} else {
		Expect(prov.Init(params)).To(BeFalse())
	}
	return prov
}

var _ = Describe("Static IP Provider", func() {
	It("Initialize provider with incorrect json format", func() {
		ipRangeHelper(`"test":"172.16.1.1-172.16.1.5", "prod":"172.16.1.50-172.16.1.55"`, false)
	})
	It("Initialize provider with incorrect ip range parameters", func() {
		ipRangeHelper(`{"test":"172.16.1.1/24", "prod":"172.16.1.50-172.16.1.55"}`, false)
	})
	It("Initialize provider with invalid start ip address", func() {
		ipRangeHelper(`{"test":"172.16.1.300-172.16.1.3"}`, false)
	})
	It("Initialize provider with invalid end ip address", func() {
		ipRangeHelper(`{"test":"172.16.1.1-172.16.1.300"}`, false)
	})
	It("Initialize provider with same start and end ip address", func() {
		ipRangeHelper(`{"test":"172.16.1.1-172.16.1.1"}`, true)
	})
	It("Changing IP Range for an IPAM Label in store", func() {
		params := Params{
			Range: `{"test":"172.16.1.1-172.16.1.5", "prod":"172.16.1.50-172.16.1.55"}`,
		}
		ipamMap := make(map[string]string)
		ipamMap["dev"] = "192.168.1.10-192.168.1.15"
		ipamMap["test"] = "192.168.1.1-192.168.1.5"
		ipamMap["prod"] = "172.16.1.50-172.16.1.55"
		store := mock.NewMockStore(mock.MockData{IPAMLabelMap: ipamMap})
		prov := &IPAMProvider{
			store:      store,
			ipamLabels: make(map[string]bool),
		}
		Expect(prov.Init(params)).To(BeTrue())
		Expect(store.Data.CleanUpFlag).To(BeTrue())
	})
	It("Test store functions", func() {
		store := mock.NewMockStore(mock.MockData{
			IPAMLabelMap: make(map[string]string),
			CleanUpFlag:  false,
			Hostdata:     make(map[string]string),
			IpList:       []string{"10.0.0.1", "10.0.0.2"},
			LabelData:    make(map[string]string),
		})
		ipamMap := make(map[string]bool)
		ipamMap["dev"] = true
		prov := &IPAMProvider{
			store:      store,
			ipamLabels: ipamMap,
		}
		prov.CreateARecord("foo.com", "10.1.1.1")
		_, ok := store.Data.Hostdata["foo.com"]
		Expect(ok).To(BeTrue())
		// Delete A Record
		prov.DeleteARecord("foo.com", "10.1.1.1")
		_, ok = store.Data.Hostdata["foo.com"]
		Expect(ok).To(BeFalse())
		prov.AllocateNextIPAddress("dev", "foo.com")
		ip, status := store.Data.LabelData["dev"]
		Expect(status).To(BeTrue())
		// Get the ipaddress from valid label
		Expect(prov.GetIPAddressFromReference("dev", "foo.com")).To(Equal(store.Data.LabelData["dev"]))
		// Releasing the ip address
		prov.ReleaseAddr(ip)
		_, ok = store.Data.LabelData["dev"]
		Expect(ok).To(BeFalse())
		// Allocate ip address from invalid label
		prov.AllocateNextIPAddress("invalid", "invalid")
		_, ok = store.Data.LabelData["invalid"]
		Expect(ok).To(BeFalse())
		// get the ipaddress from invalid label
		Expect(prov.GetIPAddressFromReference("invalid", "")).To(Equal(""))
	})
	It("Initialize provider with multiple ranges for same label", func() {
		ipRangeHelper(`{"test":"172.16.1.1-172.16.1.10,172.16.1.21-172.16.1.30"}`, true)
	})
	It("Initialize provider with a combination of label with multiple ranges and lable with single range", func() {
		ipRangeHelper(`{"test":"172.16.1.1-172.16.1.10,172.16.1.21-172.16.1.30", "prod":"172.16.1.50-172.16.1.55"}`, true)
	})
	It("Initialize provider with a combination of label with multiple ranges and lable with single range", func() {
		ipRangeHelper(`{"default":"172.16.2.50-172.16.2.55","test":"172.16.1.1-172.16.1.10,172.16.1.21-172.16.1.30", "prod":"172.16.1.50-172.16.1.55"}`, true)
	})
})
