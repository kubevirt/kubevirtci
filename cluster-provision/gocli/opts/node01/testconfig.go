package node01

import kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient) {
	cmds := []string{
		`if [ -f /home/vagrant/enable_audit ]; then echo '` + string(advAudit) + `' | tee /etc/kubernetes/audit/adv-audit.yaml > /dev/null; fi`,
		`timeout=30; interval=5; while ! hostnamectl | grep Transient; do echo "Waiting for dhclient to set the hostname from dnsmasq"; sleep $interval; timeout=$((timeout - interval)); [ $timeout -le 0 ] && exit 1; done`,
		`mkdir -p /etc/crio/crio.conf.d`,
		`[ -f /sys/fs/cgroup/cgroup.controllers ] && mkdir -p /etc/crio/crio.conf.d && echo '` + string(cgroupv2) + `' |  tee /etc/crio/crio.conf.d/00-cgroupv2.conf > /dev/null &&  sed -i 's/--cgroup-driver=systemd/--cgroup-driver=cgroupfs/' /etc/sysconfig/kubelet && systemctl stop kubelet && systemctl restart crio && systemctl start kubelet`,
		"while [[ $(systemctl status crio | grep -c active) -eq 0 ]]; do sleep 2; done",
		"swapoff -a",
		"until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done",
		`kubeadm init --config /etc/kubernetes/kubeadm.conf -v5`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p "$(cat /provision/kubeadm-patches/add-security-context-deployment-patch.yaml)"`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /provision/cni.yaml`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/control-plane:NoSchedule-`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers; kubectl_rc=$?; retry_counter=0; while [[ $retry_counter -lt 20 && $kubectl_rc -ne 0 ]]; do sleep 10; echo "Waiting for api server to be available...";  kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers; kubectl_rc=$?; retry_counter=$((retry_counter + 1)); done`,
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf version",
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /provision/local-volume.yaml`,
		"mkdir -p /var/lib/rook",
		"chcon -t container_file_t /var/lib/rook",
	}
	for _, cmd := range cmds {
		sshClient.EXPECT().Command(cmd)
	}
}
