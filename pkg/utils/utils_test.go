package utils

import (
	. "github.com/onsi/ginkgo/v2"
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

	It("Ipv4ToPaddedString pads each octet to 3 digits", func() {
		Expect(Ipv4ToPaddedString("10.65.82.0")).To(Equal("010.065.082.000"))
		Expect(Ipv4ToPaddedString("10.65.83.0")).To(Equal("010.065.083.000"))
		Expect(Ipv4ToPaddedString("0.0.0.0")).To(Equal("000.000.000.000"))
		Expect(Ipv4ToPaddedString("10.65.82.1")).To(Equal("010.065.082.001"))
		Expect(Ipv4ToPaddedString("10.65.82.10")).To(Equal("010.065.082.010"))
		Expect(Ipv4ToPaddedString("10.65.82.20")).To(Equal("010.065.082.020"))
		Expect(Ipv4ToPaddedString("192.168.1.100")).To(Equal("192.168.001.100"))
		Expect(Ipv4ToPaddedString("255.255.255.255")).To(Equal("255.255.255.255"))
	})

	It("PaddedStringToIPV4 un-pads zero octets correctly (Ticket 000222)", func() {
		// The .0-boundary cases that were previously truncated.
		Expect(PaddedStringToIPV4("010.065.082.000")).To(Equal("10.65.82.0"))
		Expect(PaddedStringToIPV4("010.065.083.000")).To(Equal("10.65.83.0"))
		Expect(PaddedStringToIPV4("000.000.000.000")).To(Equal("0.0.0.0"))
		// Any zero octet, not just the last one.
		Expect(PaddedStringToIPV4("010.000.082.005")).To(Equal("10.0.82.5"))
		Expect(PaddedStringToIPV4("000.065.082.001")).To(Equal("0.65.82.1"))
	})

	It("PaddedStringToIPV4 un-pads non-zero octets correctly", func() {
		Expect(PaddedStringToIPV4("010.065.082.001")).To(Equal("10.65.82.1"))
		Expect(PaddedStringToIPV4("010.065.082.010")).To(Equal("10.65.82.10"))
		Expect(PaddedStringToIPV4("010.065.082.020")).To(Equal("10.65.82.20"))
		Expect(PaddedStringToIPV4("192.168.001.100")).To(Equal("192.168.1.100"))
		Expect(PaddedStringToIPV4("255.255.255.255")).To(Equal("255.255.255.255"))
	})

	It("PaddedStringToIPV4 returns input unchanged for non-IPv4 input", func() {
		// IPv6 addresses do not have 4 dot-separated parts, so they pass through.
		Expect(PaddedStringToIPV4("2000::ff23")).To(Equal("2000::ff23"))
		Expect(PaddedStringToIPV4("2001:db8:85a3::8a2e:370:7334")).To(Equal("2001:db8:85a3::8a2e:370:7334"))
		Expect(PaddedStringToIPV4("::1")).To(Equal("::1"))
		// Malformed / non-numeric octets are returned unchanged rather than corrupted.
		Expect(PaddedStringToIPV4("10.65.82")).To(Equal("10.65.82"))
		Expect(PaddedStringToIPV4("10.65.82.abc")).To(Equal("10.65.82.abc"))
		Expect(PaddedStringToIPV4("")).To(Equal(""))
	})

	It("IPv4 pad/un-pad round-trips without loss", func() {
		ipv4Addrs := []string{
			"10.65.82.0",
			"10.65.83.0",
			"0.0.0.0",
			"255.255.255.255",
			"10.65.82.1",
			"10.65.82.10",
			"10.65.82.20",
			"192.168.1.100",
			"172.16.1.1",
			"10.0.0.0",
		}
		for _, ip := range ipv4Addrs {
			Expect(PaddedStringToIPV4(Ipv4ToPaddedString(ip))).To(Equal(ip), "round-trip failed for %s", ip)
		}
	})

	It("IPv6 addresses are not padded and survive the un-pad path", func() {
		ipv6Addrs := []string{
			"2000::ff23",
			"2001:db8:85a3::8a2e:370:7334",
			"::1",
			"fe80::1",
			"2001:0db8:0000:0000:0000:0000:0000:0001",
		}
		for _, ip := range ipv6Addrs {
			// IPv6 addresses must never be altered by the IPv4 un-pad routine.
			Expect(PaddedStringToIPV4(ip)).To(Equal(ip), "IPv6 address altered: %s", ip)
			Expect(IsIPV6Addr(ip)).To(BeTrue(), "expected valid IPv6: %s", ip)
		}
	})
})

func TestPaddedStringToIPV4_ZeroOctets(t *testing.T) {
	cases := map[string]string{
		"010.065.082.000": "10.65.82.0", // .0 boundary (Ticket 000222)
		"010.065.083.000": "10.65.83.0",
		"000.000.000.000": "0.0.0.0",
		"010.065.082.001": "10.65.82.1",
		"010.065.082.010": "10.65.82.10",
		"192.168.001.100": "192.168.1.100",
	}
	for padded, want := range cases {
		if got := PaddedStringToIPV4(padded); got != want {
			t.Errorf("PaddedStringToIPV4(%q) = %q, want %q", padded, got, want)
		}
	}
}

func TestPaddedRoundTrip(t *testing.T) {
	for _, ip := range []string{"10.65.82.0", "0.0.0.0", "255.255.255.255", "10.65.82.10"} {
		if got := PaddedStringToIPV4(Ipv4ToPaddedString(ip)); got != ip {
			t.Errorf("round-trip %q = %q", ip, got)
		}
	}
}

func TestPaddedStringToIPV4_IPv6PassThrough(t *testing.T) {
	// IPv6 addresses (and other non-4-octet strings) must be returned unchanged.
	for _, ip := range []string{
		"2000::ff23",
		"2001:db8:85a3::8a2e:370:7334",
		"::1",
		"fe80::1",
		"10.65.82",       // too few octets
		"10.65.82.abc",   // non-numeric octet
		"",               // empty
	} {
		if got := PaddedStringToIPV4(ip); got != ip {
			t.Errorf("PaddedStringToIPV4(%q) = %q, want unchanged", ip, got)
		}
	}
}
