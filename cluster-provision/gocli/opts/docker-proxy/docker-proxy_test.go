package dockerproxy

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestDockerProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DockerProxyOpt Suite")
}

var _ = Describe("TestDockerProxy", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *dockerProxyOpt
		cmds      []string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewDockerProxyOpt(sshClient, "test-proxy")
		script := strings.ReplaceAll(string(override), "$PROXY", opt.proxy)
		cmds = []string{
			"curl " + opt.proxy + "/ca.crt > /etc/pki/ca-trust/source/anchors/docker_registry_proxy.crt",
			"update-ca-trust",
			"mkdir -p /etc/systemd/system/crio.service.d",
			"echo '" + script + "' | sudo tee /etc/systemd/system/crio.service.d/override.conf > /dev/null",
			"systemctl daemon-reload",
			"systemctl restart crio.service",
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute DockerProxyOpt successfully", func() {
		for _, cmd := range cmds {
			sshClient.EXPECT().Command(cmd)
		}

		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
