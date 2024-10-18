package labelnodes

import (
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient, label string) {
	sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf label node -l " + label + " node-role.kubernetes.io/worker=''")
}
