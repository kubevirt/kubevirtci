package rookceph

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestNfsCsiOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CephOpt Suite")
}

var _ = Describe("CephOpt", func() {
	var (
		k8sClient k8s.K8sDynamicClient
		opt       *cephOpt
	)

	BeforeEach(func() {
		r := k8s.NewReactorConfig("create", "cephblockpools", CephReactor)
		k8sClient = k8s.NewTestClient(r)
		opt = NewCephOpt(k8sClient)
	})

	It("should execute NfsCsiOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
