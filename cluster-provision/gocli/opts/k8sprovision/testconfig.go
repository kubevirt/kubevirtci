package k8sprovision

import (
	"fmt"
	"strings"

	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient, version string, slim bool) {
	crio, _ := f.ReadFile("conf/crio-yum.repo")
	registries, _ := f.ReadFile("conf/registries.conf")
	storage, _ := f.ReadFile("conf/storage.conf")
	k8sRepo, _ := f.ReadFile("conf/kubernetes.repo")
	cniPatch, _ := f.ReadFile("conf/cni.diff")
	cniV6Patch, _ := f.ReadFile("conf/cni_ipv6.diff")
	k8sConf, _ := f.ReadFile("conf/k8s.conf")
	calico, _ := f.ReadFile("conf/001-calico.conf")
	dhclient, _ := f.ReadFile("conf/002-dhclient.conf")
	secContextPatch, _ := patchFs.ReadFile("patches/add-security-context-deployment-patch.yaml")
	etcdPatch, _ := patchFs.ReadFile("patches/etcd.yaml")
	apiServerPatch, _ := patchFs.ReadFile("patches/kube-apiserver.yaml")
	controllerManagerPatch, _ := patchFs.ReadFile("patches/kube-controller-manager.yaml")
	schedulerPatch, _ := patchFs.ReadFile("patches/kube-scheduler.yaml")

	packagesVersion := "1.30"

	advAudit, _ := f.ReadFile("conf/adv-audit.yaml")
	psa, _ := f.ReadFile("conf/psa.yaml")
	kubeAdm, _ := f.ReadFile("conf/kubeadm.conf")
	kubeAdm6, _ := f.ReadFile("conf/kubeadm_ipv6.conf")

	k8sMinor := strings.Split(version, ".")[1]
	k8sRepoWithVersion := strings.Replace(string(k8sRepo), "VERSION", k8sMinor, 1)
	kubeAdmConf := strings.Replace(string(kubeAdm), "VERSION", version, 1)
	kubeAdm6Conf := strings.Replace(string(kubeAdm6), "VERSION", version, 1)

	cmds := []string{
		"echo '" + string(crio) + "' | tee /etc/yum.repos.d/devel_kubic_libcontainers_stable_cri-o_v1.28.repo >> /dev/null",
		"dnf install -y cri-o",
		"echo '" + string(registries) + "' | tee /etc/containers/registries.conf >> /dev/null",
		"echo '" + string(storage) + "' | tee /etc/containers/storage.conf >> /dev/null",
		"systemctl restart crio",
		"systemctl enable --now crio",
		"echo '" + k8sRepoWithVersion + "' | tee /etc/yum.repos.d/kubernetes.repo >> /dev/null",
		fmt.Sprintf("dnf install --skip-broken --nobest --nogpgcheck --disableexcludes=kubernetes -y kubectl-%[1]s kubeadm-%[1]s kubelet-%[1]s kubernetes-cni", packagesVersion),
		"kubeadm config images pull --kubernetes-version " + version,
		`image_regex='([a-z0-9\_\.]+[/-]?)+(@sha256)?:[a-z0-9\_\.\-]+' image_regex_w_double_quotes='"?'"${image_regex}"'"?' find /tmp -type f -name '*.yaml' -print0 | xargs -0 grep -iE '(image|value): '"${image_regex_w_double_quotes}" > /tmp/images`,
	}

	for _, cmd := range cmds {
		sshClient.EXPECT().Command(cmd)
	}

	sshClient.EXPECT().CommandWithNoStdOut(`image_regex='([a-z0-9\_\.]+[/-]?)+(@sha256)?:[a-z0-9\_\.\-]+' && image_regex_w_double_quotes='"?'"${image_regex}"'"?' && grep -ioE "${image_regex_w_double_quotes}" /tmp/images`).Return("nginx:latest", nil)

	cmds = []string{
		"mkdir /provision",
		"yum install -y patch || true",
		"dnf install -y patch || true",
		"cp /tmp/cni.do-not-change.yaml /provision/cni.yaml",
		"mv /tmp/cni.do-not-change.yaml /provision/cni_ipv6.yaml",
		"echo '" + string(cniPatch) + "' | tee /tmp/cni_patch.diff >> /dev/null",
		"echo '" + string(cniV6Patch) + "' | tee /tmp/cni_v6_patch.diff >> /dev/null",
		"patch /provision/cni.yaml /tmp/cni_patch.diff",
		"patch /provision/cni_ipv6.yaml /tmp/cni_v6_patch.diff",
		"cp /tmp/local-volume.yaml /provision/local-volume.yaml",
		`echo "vm.unprivileged_userfaultfd = 1" > /etc/sysctl.d/enable-userfaultfd.conf`,
		"modprobe bridge",
		"modprobe overlay",
		"modprobe br_netfilter",
		"echo '" + string(k8sConf) + "' | tee /etc/sysctl.d/k8s.conf >> /dev/null",
		"sysctl --system",
		"echo bridge >> /etc/modules-load.d/k8s.conf",
		"echo br_netfilter >> /etc/modules-load.d/k8s.conf",
		"echo overlay >> /etc/modules-load.d/k8s.conf",
		"rm -f /etc/cni/net.d/*",
		"systemctl daemon-reload",
		"systemctl enable crio kubelet --now",
		"echo '" + string(calico) + "' | tee /etc/NetworkManager/conf.d/001-calico.conf >> /dev/null",
		"echo '" + string(dhclient) + "' | tee /etc/NetworkManager/conf.d/002-dhclient.conf >> /dev/null",
		`echo "net.netfilter.nf_conntrack_max=1000000" >> /etc/sysctl.conf`,
		"sysctl --system",
		"systemctl restart NetworkManager",
		`nmcli connection modify "System eth0" ipv6.method auto ipv6.addr-gen-mode eui64`,
		`nmcli connection up "System eth0"`,
		"sysctl --system",
		"echo bridge >> /etc/modules-load.d/k8s.conf",
		"echo br_netfilter >> /etc/modules-load.d/k8s.conf",
		"echo overlay >> /etc/modules-load.d/k8s.conf",
		"mkdir -p /provision/kubeadm-patches",
		"echo '" + string(secContextPatch) + "' | tee /provision/kubeadm-patches/add-security-context-deployment-patch.yaml >> /dev/null",
		"echo '" + string(etcdPatch) + "' | tee /provision/kubeadm-patches/etcd.yaml >> /dev/null",
		"echo '" + string(apiServerPatch) + "' | tee /provision/kubeadm-patches/kube-apiserver.yaml >> /dev/null",
		"echo '" + string(controllerManagerPatch) + "' | tee /provision/kubeadm-patches/kube-controller-manager.yaml >> /dev/null",
		"echo '" + string(schedulerPatch) + "' | tee /provision/kubeadm-patches/kube-scheduler.yaml >> /dev/null",
		"mkdir /etc/kubernetes/audit",
		"echo '" + string(advAudit) + "' | tee /etc/kubernetes/audit/adv-audit.yaml >> /dev/null",
		"echo '" + string(psa) + "' | tee /etc/kubernetes/psa.yaml >> /dev/null",
		"echo '" + kubeAdmConf + "' | tee /etc/kubernetes/kubeadm.conf >> /dev/null",
		"echo '" + kubeAdm6Conf + "' | tee /etc/kubernetes/kubeadm_ipv6.conf >> /dev/null",
		"until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done",
		"swapoff -a",
		"systemctl restart kubelet",
		"kubeadm init --config /etc/kubernetes/kubeadm.conf -v5",
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p '" + string(secContextPatch) + "'",
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /provision/cni.yaml",
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pods --all -n kube-system --timeout=300s",
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system",
		"kubeadm reset --force",
		"mkdir -p /var/provision/kubevirt.io/tests",
		"chcon -t container_file_t /var/provision/kubevirt.io/tests",
		`echo "tmpfs /var/provision/kubevirt.io/tests tmpfs rw,context=system_u:object_r:container_file_t:s0 0 1" >> /etc/fstab`,
		"rm -f /etc/sysconfig/network-scripts/ifcfg-*",
		"nmcli connection add con-name eth0 ifname eth0 type ethernet",
		"rm -f /etc/machine-id ; touch /etc/machine-id",
	}

	for _, cmd := range cmds {
		sshClient.EXPECT().Command(cmd)
	}
}
