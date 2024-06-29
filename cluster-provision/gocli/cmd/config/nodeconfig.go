package config

// NodeLinuxConfig type is a holder for all the config params that a node can have for its linux system
type NodeLinuxConfig struct {
	NodeIdx      int
	K8sVersion   string
	FipsEnabled  bool
	DockerProxy  string
	EtcdInMemory bool
	EtcdSize     string
	SingleStack  bool
	EnableAudit  bool
	GpuAddress   string
	Realtime     bool
	PSA          bool
}

// NodeK8sConfig type is a holder for all the config k8s options for kubevirt cluster
type NodeK8sConfig struct {
	Ceph         bool
	Prometheus   bool
	Alertmanager bool
	Grafana      bool
	Istio        bool
	NfsCsi       bool
}

func NewNodeK8sConfig(ceph, prometheus, alertmanager, grafana, istio, nfsCsi bool) *NodeK8sConfig {
	return &NodeK8sConfig{
		Ceph:         ceph,
		Prometheus:   prometheus,
		Alertmanager: alertmanager,
		Grafana:      grafana,
		Istio:        istio,
		NfsCsi:       nfsCsi,
	}
}

func NewNodeLinuxConfig(nodeIdx int, k8sVersion, dockerProxy, etcdSize, gpuAddress string,
	fipsEnabled, etcdInMemory, singleStack, enableAudit, realtime, psa bool) *NodeLinuxConfig {
	return &NodeLinuxConfig{
		NodeIdx:      nodeIdx,
		K8sVersion:   k8sVersion,
		FipsEnabled:  fipsEnabled,
		DockerProxy:  dockerProxy,
		EtcdInMemory: etcdInMemory,
		EtcdSize:     etcdSize,
		SingleStack:  singleStack,
		EnableAudit:  enableAudit,
		GpuAddress:   gpuAddress,
		Realtime:     realtime,
		PSA:          psa,
	}
}
