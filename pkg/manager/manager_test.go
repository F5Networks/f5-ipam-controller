package manager

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Static IP Provider", func() {
	It("New Manger test", func() {
		params := Params{InfobloxProvider,
			IPAMManagerParams{`"test":"172.16.1.1-172.16.1.5", "prod":"172.16.1.50-172.16.1.55"`},
			InfobloxParams{"localhost",
				"2.2.6",
				"6443",
				"admin",
				"infoblox",
				"{\"Dev\" :{\"cidr\": \"172.16.4.0/24\"},\"Test\" :{\"cidr\": \"172.16.5.0/24\"}}",
				"default",
				"false"}}
		_, err := NewManager(params)
		Expect(err).NotTo(BeEquivalentTo(nil))
		params.Provider = F5IPAMProvider
		_, err = NewManager(params)
		Expect(err).NotTo(BeEquivalentTo(nil))
		params.Provider = "default"
		_, err = NewManager(params)
		Expect(err).NotTo(BeEquivalentTo(nil))
	})
})
