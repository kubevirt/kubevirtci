package cmd

import (
	"fmt"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	nodesconfig "kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/config"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

type TestSuite struct {
	suite.Suite
	sshClient *kubevirtcimocks.MockSSHClient
}

func (ts *TestSuite) SetupTest() {
	ts.sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(ts.T()))
}

func (ts *TestSuite) TearDownTest() {
	ts.sshClient = nil
}

func (ts *TestSuite) TestProvisionNode() {
	n := nodesconfig.NewNodeLinuxConfig(1, "k8s-1.30", "", "512M", "", false, true, true, true, true, true)
	cmds := []string{
		"mkdir -p /var/lib/etcd",
		"test -d /var/lib/etcd",
		fmt.Sprintf("mount -t tmpfs -o size=%s tmpfs /var/lib/etcd", n.EtcdSize),
		"df -h /var/lib/etcd",
		"/scripts/realtime.sh",
		"-s -- --vendor 8086:2668 < /scripts/bind_device_to_vfio.sh",
		"-s -- --vendor 8086:2415 < /scripts/bind_device_to_vfio.sh",
		"touch /home/vagrant/single_stack",
		"touch /home/vagrant/enable_audit",
		"/scripts/psa.sh",
		"/scripts/node01.sh",
	}

	for _, cmd := range cmds {
		ts.sshClient.EXPECT().SSH(cmd).Return(nil)
	}

	err := provisionNode(ts.sshClient, n)
	ts.NoError(err)
}

func (ts *TestSuite) TestProvisionNodeK8sOpts() {
	n := nodesconfig.NewNodeK8sConfig(true, true, true, true, true, true)
	cmds := []string{
		"/scripts/rook-ceph.sh",
		"/scripts/nfs-csi.sh",
		"/scripts/istio.sh",
		"-s -- --alertmanager true --grafana true  < /scripts/prometheus.sh",
	}

	for _, cmd := range cmds {
		ts.sshClient.EXPECT().SSH(cmd).Return(nil)
	}

	err := provisionK8sOptions(ts.sshClient, n)
	ts.NoError(err)
}
