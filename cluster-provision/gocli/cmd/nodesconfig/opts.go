package nodesconfig

type LinuxConfigFunc func(n *NodeLinuxConfig)

type K8sConfigFunc func(n *NodeK8sConfig)

func WithNodeIdx(nodeIdx int) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.NodeIdx = nodeIdx
	}
}

func WithFipsEnabled(fipsEnabled bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.FipsEnabled = fipsEnabled
	}
}

func WithDockerProxy(dockerProxy string) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.DockerProxy = dockerProxy
	}
}

func WithEtcdInMemory(etcdInMemory bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.EtcdInMemory = etcdInMemory
	}
}

func WithEtcdSize(etcdSize string) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.EtcdSize = etcdSize
	}
}

func WithSingleStack(singleStack bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.SingleStack = singleStack
	}
}

func WithFlannel(flannel bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.Flannel = flannel
	}
}

func WithKindnet(kindnet bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.Kindnet = kindnet
	}
}

func WithNoEtcdFsync(noEtcdFsync bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.NoEtcdFsync = noEtcdFsync
	}
}

func WithEnableAudit(enableAudit bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.EnableAudit = enableAudit
	}
}

func WithGpuAddress(gpuAddress string) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.GpuAddress = gpuAddress
	}
}

func WithRealtime(realtime bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.Realtime = realtime
	}
}

func WithPSA(psa bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.PSA = psa
	}
}

func WithKsm(ksm bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.KsmEnabled = ksm
	}
}

func WithSwap(swap bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.SwapEnabled = swap
	}
}

func WithKsmEnabled(ksmEnabled bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.KsmEnabled = ksmEnabled
	}
}

func WithSwapEnabled(swapEnabled bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.SwapEnabled = swapEnabled
	}
}

func WithKsmPageCount(ksmPageCount int) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.KsmPageCount = ksmPageCount
	}
}

func WithKsmScanInterval(ksmScanInterval int) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.KsmScanInterval = ksmScanInterval
	}
}

func WithSwapiness(swapiness int) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.Swappiness = swapiness
	}
}

func WithUnlimitedSwap(unlimitedSwap bool) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.UnlimitedSwap = unlimitedSwap
	}
}

func WithSwapSize(swapSize int) LinuxConfigFunc {
	return func(n *NodeLinuxConfig) {
		n.SwapSize = swapSize
	}
}

func WithCeph(ceph bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.Ceph = ceph
	}
}

func WithPrometheus(prometheus bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.Prometheus = prometheus
	}
}

func WithAlertmanager(alertmanager bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.Alertmanager = alertmanager
	}
}

func WithGrafana(grafana bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.Grafana = grafana
	}
}

func WithIstio(istio bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.Istio = istio
	}
}

func WithNfsCsi(nfsCsi bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.NfsCsi = nfsCsi
	}
}

func WithCnao(cnao bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.CNAO = cnao
	}
}

// Skips creation of the CNAO custom resource. Just installs the CRD and the operator
func WithCNAOSkipCR(skip bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.CNAOSkipCR = skip
	}
}

// Whether or no to deploy the dynamic networks controller through CNAO
func WithDNC(dnc bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.DNC = dnc
	}
}

// If enabled. Multus V3 will be deployed standalone separate from CNAO.
// Multus V4 that gets deployed with CNAO will be skipped in case CNAO is enabled
func WithMultus(multus bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.Multus = multus
	}
}

func WithCdi(cdi bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.CDI = cdi
	}
}

func WithCdiVersion(cdiVersion string) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.CDIVersion = cdiVersion
	}
}

func WithAAQ(aaq bool) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.AAQ = aaq
	}
}

func WithAAQVersion(aaqVersion string) K8sConfigFunc {
	return func(n *NodeK8sConfig) {
		n.AAQVersion = aaqVersion
	}
}
