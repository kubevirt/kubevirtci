package k8sprovision

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed conf/*
var f embed.FS

//go:embed patches/*
var patchFs embed.FS

type k8sProvisioner struct {
	version   string
	slim      bool
	sshClient libssh.Client
}

func NewK8sProvisioner(sshClient libssh.Client, version string, slim bool) *k8sProvisioner {
	return &k8sProvisioner{
		version:   version,
		slim:      slim,
		sshClient: sshClient,
	}
}

func (k *k8sProvisioner) Exec() error {
	crio, err := f.ReadFile("conf/crio-yum.repo")
	if err != nil {
		return err
	}

	registries, err := f.ReadFile("conf/registries.conf")
	if err != nil {
		return err
	}

	storage, err := f.ReadFile("conf/storage.conf")
	if err != nil {
		return err
	}

	mult, err := f.ReadFile("conf/multus.yaml")
	if err != nil {
		return err
	}

	k8sRepo, err := f.ReadFile("conf/kubernetes.repo")
	if err != nil {
		return err
	}

	cniPatch, err := f.ReadFile("conf/cni.diff")
	if err != nil {
		return err
	}

	cniV6Patch, err := f.ReadFile("conf/cni_ipv6.diff")
	if err != nil {
		return err
	}

	k8sConf, err := f.ReadFile("conf/k8s.conf")
	if err != nil {
		return err
	}

	calico, err := f.ReadFile("conf/001-calico.conf")
	if err != nil {
		return err
	}

	dhclient, err := f.ReadFile("conf/002-dhclient.conf")
	if err != nil {
		return err
	}

	secContextPatch, err := patchFs.ReadFile("patches/add-security-context-deployment-patch.yaml")
	if err != nil {
		return err
	}

	etcdPatch, err := patchFs.ReadFile("patches/etcd.yaml")
	if err != nil {
		return err
	}

	apiServerPatch, err := patchFs.ReadFile("patches/kube-apiserver.yaml")
	if err != nil {
		return err
	}

	controllerManagerPatch, err := patchFs.ReadFile("patches/kube-controller-manager.yaml")
	if err != nil {
		return err
	}

	schedulerPatch, err := patchFs.ReadFile("patches/kube-scheduler.yaml")
	if err != nil {
		return err
	}

	packagesVersion, err := k.getPackagesVersion()
	if err != nil {
		return err
	}

	advAudit, err := f.ReadFile("conf/adv-audit.yaml")
	if err != nil {
		return err
	}

	psa, err := f.ReadFile("conf/psa.yaml")
	if err != nil {
		return err
	}

	kubeAdm, err := f.ReadFile("conf/kubeadm.conf")
	if err != nil {
		return err
	}

	kubeAdm6, err := f.ReadFile("conf/kubeadm_ipv6.conf")
	if err != nil {
		return err
	}

	k8sMinor := strings.Split(k.version, ".")[1]

	crioWithVersion := strings.Replace(string(crio), "VERSION", k8sMinor, -1)
	k8sRepoWithVersion := strings.Replace(string(k8sRepo), "VERSION", k8sMinor, -1)
	kubeAdmConf := strings.Replace(string(kubeAdm), "VERSION", k.version, -1)
	kubeAdm6Conf := strings.Replace(string(kubeAdm6), "VERSION", k.version, -1)

	cmds := []string{
		"echo '" + crioWithVersion + "' | tee /etc/yum.repos.d/devel_kubic_libcontainers_stable_cri-o_v1." + k8sMinor + ".repo >> /dev/null",
		"dnf install -y cri-o",
		"echo '" + string(registries) + "' | tee /etc/containers/registries.conf >> /dev/null",
		"echo '" + string(storage) + "' | tee /etc/containers/storage.conf >> /dev/null",
		"systemctl restart crio",
		"systemctl enable --now crio",
		"echo '" + k8sRepoWithVersion + "' | tee /etc/yum.repos.d/kubernetes.repo >> /dev/null",
		fmt.Sprintf("dnf install --skip-broken --nobest --nogpgcheck --disableexcludes=kubernetes -y kubectl-%[1]s kubeadm-%[1]s kubelet-%[1]s kubernetes-cni", packagesVersion),
		"kubeadm config images pull --kubernetes-version " + k.version,
		`image_regex='([a-z0-9\_\.]+[/-]?)+(@sha256)?:[a-z0-9\_\.\-]+' image_regex_w_double_quotes='"?'"${image_regex}"'"?' find /tmp -type f -name '*.yaml' -print0 | xargs -0 grep -iE '(image|value): '"${image_regex_w_double_quotes}" > /tmp/images`,
	}

	for _, cmd := range cmds {
		if err := k.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	images, err := k.sshClient.CommandWithNoStdOut(`image_regex='([a-z0-9\_\.]+[/-]?)+(@sha256)?:[a-z0-9\_\.\-]+' && image_regex_w_double_quotes='"?'"${image_regex}"'"?' && grep -ioE "${image_regex_w_double_quotes}" /tmp/images`)
	if err != nil {
		return err
	}

	if !k.slim {
		imagesList := strings.Split(images, "\n")
		for _, image := range imagesList {
			err := k.pullImageRetry(image)
			if err != nil {
				logrus.Infof("Failed to pull image: %s, it will not be available offline", image)
			}
		}

		extraImg, err := f.ReadFile("conf/extra-images")
		if err != nil {
			return err
		}

		imagesList = strings.Split(string(extraImg), "\n")
		for _, image := range imagesList {
			err := k.pullImageRetry(image)
			if err != nil {
				logrus.Infof("Failed to pull image: %s, it will not be available offline", image)
			}
		}
	}

	cmds = []string{
		"echo '" + string(mult) + "' | tee /opt/multus.yaml >> /dev/null",
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
		"echo 'KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice  --fail-swap-on=false --kubelet-cgroups=/systemd/system.slice' >> /etc/sysconfig/kubelet",
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
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /opt/multus.yaml",
		"while ! kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system ds/kube-multus-ds --timeout=200s; do kubectl --kubeconfig=/etc/kubernetes/admin.conf describe pods -n kube-system -l name=kube-multus-ds; sleep 5; done",
		"kubeadm reset --force",
		`for i in {1..10}; do mkdir -p /var/local/kubevirt-storage/local-volume/disk${i} && mkdir -p /mnt/local-storage/local/disk${i} && echo "/var/local/kubevirt-storage/local-volume/disk${i} /mnt/local-storage/local/disk${i} none defaults,bind 0 0" >> /etc/fstab; done`,
		"chmod -R 777 /var/local/kubevirt-storage/local-volume",
		"chcon -R unconfined_u:object_r:svirt_sandbox_file_t:s0 /mnt/local-storage/",
		"mkdir -p /var/provision/kubevirt.io/tests",
		"chcon -t container_file_t /var/provision/kubevirt.io/tests",
		`echo "tmpfs /var/provision/kubevirt.io/tests tmpfs rw,context=system_u:object_r:container_file_t:s0 0 1" >> /etc/fstab`,
		"cp -rf /tmp/kwok /opt/",
		"rm -f /etc/sysconfig/network-scripts/ifcfg-*",
		"nmcli connection add con-name eth0 ifname eth0 type ethernet",
		"rm -f /etc/machine-id ; touch /etc/machine-id",
	}

	for _, cmd := range cmds {
		if err := k.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}

func (k *k8sProvisioner) pullImageRetry(image string) error {
	maxRetries := 5
	downloaded := false

	for i := 0; i < maxRetries; i++ {
		if err := k.sshClient.Command("crictl pull " + image); err != nil {
			logrus.Infof("Attempt [%d]: Failed to download image %s: %s, sleeping 3 seconds and trying again", i+1, image, err.Error())
			time.Sleep(time.Second * 3)
		} else {
			downloaded = true
			break
		}
	}

	if !downloaded {
		return fmt.Errorf("reached max retries to download for %s", image)
	}
	return nil
}

func (k *k8sProvisioner) getPackagesVersion() (string, error) {
	packagesVersion := k.version
	if strings.HasSuffix(k.version, "alpha") || strings.HasSuffix(k.version, "beta") || strings.HasSuffix(k.version, "rc") {
		k8sversion := strings.Split(k.version, ".")

		url := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/stable-%s.%s.txt", k8sversion[0], k8sversion[1])
		resp, err := http.Get(url)
		if err != nil {
			return packagesVersion, nil
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading the response body:", err)
			return packagesVersion, nil
		}
		if string(body) != "" {
			packagesVersion = strings.TrimPrefix(string(body), "v")
		}
	}
	return packagesVersion, nil
}
