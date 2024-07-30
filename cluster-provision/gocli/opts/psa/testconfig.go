package psa

import kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient) {
	psa, _ := f.ReadFile("manifests/psa.yaml")

	sshClient.EXPECT().Command("rm /etc/kubernetes/psa.yaml", true)
	sshClient.EXPECT().Command("echo '"+string(psa)+"' | sudo tee /etc/kubernetes/psa.yaml > /dev/null", true)
}
