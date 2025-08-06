package nodesconfig

// NodeLinuxConfig type holds the config params that a node can have for its linux system
type NodeLinuxConfig struct {
	NodeIdx         int
	K8sVersion      string
	FipsEnabled     bool
	DockerProxy     string
	EtcdInMemory    bool
	EtcdSize        string
	SingleStack     bool
	Flannel         bool
	Kindnet         bool
	NoEtcdFsync     bool
	EnableAudit     bool
	GpuAddress      string
	Realtime        bool
	PSA             bool
	KsmEnabled      bool
	SwapEnabled     bool
	KsmPageCount    int
	KsmScanInterval int
	Swappiness      int
	UnlimitedSwap   bool
	SwapSize        int
}

// NodeK8sConfig type holds the config k8s options for kubevirt cluster
type NodeK8sConfig struct {
	Ceph         bool
	Prometheus   bool
	Alertmanager bool
	Grafana      bool
	Istio        bool
	NfsCsi       bool
	CNAO         bool
	CNAOSkipCR   bool
	Multus       bool
	CDI          bool
	CDIVersion   string
	AAQ          bool
	AAQVersion   string
	DNC          bool
}

func NewNodeK8sConfig(confs []K8sConfigFunc) *NodeK8sConfig {
	n := &NodeK8sConfig{}

	for _, conf := range confs {
		conf(n)
	}

	return n
}

func NewNodeLinuxConfig(nodeIdx int, k8sVersion string, confs []LinuxConfigFunc) *NodeLinuxConfig {
	n := &NodeLinuxConfig{
		NodeIdx:    nodeIdx,
		K8sVersion: k8sVersion,
	}

	for _, conf := range confs {
		conf(n)
	}

	return n
}
