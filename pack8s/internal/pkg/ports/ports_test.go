package ports_test

import (
	"github.com/fromanirh/pack8s/internal/pkg/ports"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ports", func() {
	Context("ports", func() {
		It("Should test if port number is known", func() {
			knownPortNames := []string{ports.PortNameSSH, ports.PortNameSSHWorker, ports.PortNameAPI, ports.PortNameOCP, ports.PortNameOCPConsole, ports.PortNameRegistry, ports.PortNameVNC}
			for _, portName := range knownPortNames {
				res := ports.IsKnownPortName(portName)
				Expect(res).To(Equal(true))
			}
			res := ports.IsKnownPortName("IMAP")
			Expect(res).To(Equal(false))
		})

		It("Should convert port name to port number", func() {
			portMap := map[string]int{
				ports.PortNameSSH:        2201,
				ports.PortNameSSHWorker:  2202,
				ports.PortNameRegistry:   5000,
				ports.PortNameOCP:        8443,
				ports.PortNameAPI:        6443,
				ports.PortNameVNC:        5901,
				ports.PortNameOCPConsole: 443,
			}
			for portKey, portValue := range portMap {
				res, err := ports.NameToNumber(portKey)
				Expect(res).To(Equal(portValue))
				Expect(err).To(BeNil())
			}

			res, err := ports.NameToNumber("IMAP")
			Expect(res).To(Equal(0))
			Expect(err).NotTo(BeNil())

		})

		It("Should convert port number to string", func() {
			res := ports.ToStrings(443, 22)

			Expect(len(res)).To(Equal(2))
			Expect(res[0]).To(Equal("443"))
			Expect(res[1]).To(Equal("22"))
		})
	})
})
