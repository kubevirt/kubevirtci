package nfscsi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestNfsCsiOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NfsCsiOpt Suite")
}

var _ = Describe("NfsCsiOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		k8sClient k8s.K8sDynamicClient
		opt       *nfsCsiOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		r := k8s.NewReactorConfig("create", "persistentvolumeclaims", NfsCsiReactor)
		k8sClient = k8s.NewTestClient(r)
		opt = NewNfsCsiOpt(k8sClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute NfsCsiOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
