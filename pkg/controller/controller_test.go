package controller

import (
	"github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"
	"github.com/F5Networks/f5-ipam-controller/pkg/manager/mock"
	mockorch "github.com/F5Networks/f5-ipam-controller/pkg/orchestration/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Contoller Suite")
}

var _ = Describe("Static IP Provider", func() {
	mockData := mock.MockData{
		IPList: []string{"1.2.3.4", "2.3.4.5"},
	}
	mgr, _ := mock.NewMockIPAMManager(mockData)
	orcr := &mockorch.MockOrch{
		ReqChan:  make(chan ipamspec.IPAMRequest),
		RespChan: make(chan ipamspec.IPAMResponse),
	}
	stopCh := make(chan struct{})
	ctlr := NewController(Spec{
		Orchestrator: orcr,
		Manager:      mgr,
		StopCh:       stopCh,
	})
	It("check orch", func() {
		ctlr.Start()
		Expect(mockorch.StartCalled).To(BeTrue())
	})
	It("should process the ip request", func() {
		ctlr.reqChan <- ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "foo.com", IPAddr: "", Key: "", IPAMLabel: "Dev", CIDR: "", DNSView: ""}
		tmp := <-ctlr.respChan
		Expect(tmp.IPAddr).To(Equal("1.2.3.4"), "Should get an ip addresse.")
		ctlr.reqChan <- ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.DELETE, HostName: "foo.com", IPAddr: "1.2.3.4", Key: "", IPAMLabel: "Dev", CIDR: "", DNSView: ""}
		tmp2 := <-ctlr.respChan
		Expect(tmp2.IPAddr).To(Equal(""), "Should remove the ip addresse.")
		ctlr.reqChan <- ipamspec.IPAMRequest{Metadata: "", Operation: ipamspec.CREATE, HostName: "", IPAddr: "", Key: "Test", IPAMLabel: "Dev", CIDR: "", DNSView: ""}
		tmp3 := <-ctlr.respChan
		Expect(tmp3.IPAddr).To(Equal("1.2.3.4"), "Should get previous ip address only")
		// Fail A record Creation
		//mgr.SkipRecord(true)
		//ctlr.reqChan <- ipamspec.IPAMRequest{"", ipamspec.CREATE, "example.com", "", "", "Dev", "", ""}
		//tmp4 := <-ctlr.respChan
		//Expect(tmp4.IPAddr).To(Equal(""), "A record should not be created and ipaddress should be released")
	})
	It("check orch", func() {
		ctlr.Stop()
		Expect(mockorch.StopCalled).To(BeTrue())
	})
})
