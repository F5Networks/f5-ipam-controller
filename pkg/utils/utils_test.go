package utils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Utils function test", func() {
	It("Checking Random string generator", func() {
		Expect(len(RandomString(5))).To(BeEquivalentTo(5))
		Expect(len(RandomString(0))).To(BeEquivalentTo(0))
		Expect(len(RandomString(-1))).To(BeEquivalentTo(0))
	})
	It("Check IsIPAddr", func() {
		Expect(IsIPV4Addr("172.16.1.1")).To(BeTrue())
		Expect(IsIPV4Addr("300.300.300.300")).To(BeFalse())
		Expect(IsIPV4Addr("2000::ff23")).To(BeFalse())
		Expect(IsIPV6Addr("2000::ff23")).To(BeTrue())
		Expect(IsIPV6Addr("2000.ff23.jjjj")).To(BeFalse())
		Expect(IsIPV6Addr("172.16.1.1")).To(BeFalse())
		Expect(IsIPAddr("")).To(BeFalse())
		Expect(IsIPAddr("cdaskn")).To(BeFalse())
	})
})
