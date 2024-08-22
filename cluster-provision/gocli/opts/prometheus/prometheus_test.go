package prometheus

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestPrometheusOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PrometheusOpt Suite")
}

var _ = Describe("PrometheusOpt", func() {
	var (
		client k8s.K8sDynamicClient
		opt    *prometheusOpt
	)

	BeforeEach(func() {
		client = k8s.NewTestClient()
		opt = NewPrometheusOpt(client, true, true)
	})

	It("should execute PrometheusOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
