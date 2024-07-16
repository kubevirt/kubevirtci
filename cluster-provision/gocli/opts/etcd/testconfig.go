package etcdinmemory

import (
	"fmt"

	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient, size string) {
	cmds := []string{
		"mkdir -p /var/lib/etcd",
		"test -d /var/lib/etcd",
		fmt.Sprintf("mount -t tmpfs -o size=%s tmpfs /var/lib/etcd", size),
	}
	for _, cmd := range cmds {
		sshClient.EXPECT().Command(cmd, true)
	}
}
