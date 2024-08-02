package aaq

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestAaqOpt(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AaqOpt Suite")
}

var _ = ginkgo.Describe("AaqOpt", func() {
	var (
		client    k8s.K8sDynamicClient
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *aaqOpt
		ctrl      *gomock.Controller
	)

	ginkgo.BeforeEach(func() {
		client = k8s.NewTestClient()
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(ctrl)
		opt = NewAaqOpt(client, sshClient, "")
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.It("should execute without error", func() {
		sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pod --timeout=180s --all --namespace aaq")

		err := opt.Exec()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})
})
