package psa

import kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient) {
	sshClient.EXPECT().Command("rm /etc/kubernetes/psa.yaml")
	sshClient.EXPECT().Command("echo '" + string(psa) + "' | sudo tee /etc/kubernetes/psa.yaml > /dev/null")
}
