package cmd

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/nodesconfig"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

var _ = Describe("Node Provisioning", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
		sshClient = nil
	})

	Describe("ProvisionNode", func() {
		It("should execute the correct commands", func() {
			n := nodesconfig.NewNodeLinuxConfig(1, "k8s-1.30", "", "512M", "", false, true, true, true, true, true)
			cmds := []string{
				"mkdir -p /var/lib/etcd",
				"test -d /var/lib/etcd",
				fmt.Sprintf("mount -t tmpfs -o size=%s tmpfs /var/lib/etcd", n.EtcdSize),
				"df -h /var/lib/etcd",
				"/scripts/realtime.sh",
				"touch /home/vagrant/single_stack",
				"touch /home/vagrant/enable_audit",
				"-s -- --vendor 8086:2668 < /scripts/bind_device_to_vfio.sh",
				"-s -- --vendor 8086:2415 < /scripts/bind_device_to_vfio.sh",
				"/scripts/psa.sh",
				"/scripts/node01.sh",
			}

			for _, cmd := range cmds {
				sshClient.EXPECT().Command(cmd).Return(nil)
			}

			err := provisionNode(sshClient, n)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ProvisionNodeK8sOpts", func() {
		It("should execute the correct K8s option commands", func() {
			n := nodesconfig.NewNodeK8sConfig(true, true, true, true, true, true)
			cmds := []string{
				"/scripts/rook-ceph.sh",
				"/scripts/nfs-csi.sh",
				"/scripts/istio.sh",
				"-s -- --alertmanager true --grafana true  < /scripts/prometheus.sh",
			}

			for _, cmd := range cmds {
				sshClient.EXPECT().Command(cmd).Return(nil)
			}

			err := provisionK8sOptions(sshClient, n)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
