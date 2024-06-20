package providers

import (
	"github.com/docker/docker/client"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

type KubevirtProvider struct {
	Client  *k8s.K8sDynamicClient `json:"-"`
	Docker  *client.Client        `json:"-"`
	DNSMasq string                `json:"dnsmasq"`

	Version string `json:"version"`
	Image   string `json:"image"`
	Nodes   uint   `flag:"nodes" short:"n" json:"nodes"`
	Numa    uint   `flag:"numa" short:"u" json:"numa"`
	Memory  string `flag:"memory" short:"m" json:"memory"`
	CPU     uint   `flag:"cpu" short:"c" json:"cpu"`

	SecondaryNics                uint     `flag:"secondary-nics" json:"secondary_nics"`
	QemuArgs                     string   `flag:"qemu-args" json:"qemu_args"`
	KernelArgs                   string   `flag:"kernel-args" json:"kernel_args"`
	Background                   bool     `flag:"background" short:"b" json:"background"`
	Reverse                      bool     `flag:"reverse" short:"r" json:"reverse"`
	RandomPorts                  bool     `flag:"random-ports" json:"random_ports"`
	Slim                         bool     `flag:"slim" json:"slim"`
	VNCPort                      uint16   `flag:"vnc-port" json:"vnc_port"`
	HTTPPort                     uint16   `flag:"http-port" json:"http_port"`
	HTTPSPort                    uint16   `flag:"https-port" json:"https_port"`
	RegistryPort                 uint16   `flag:"registry-port" json:"registry_port"`
	OCPort                       uint16   `flag:"ocp-port" json:"ocp_port"`
	K8sPort                      uint16   `flag:"k8s-port" json:"k8s_port"`
	SSHPort                      uint16   `flag:"ssh-port" json:"ssh_port"`
	PrometheusPort               uint16   `flag:"prometheus-port" json:"prometheus_port"`
	GrafanaPort                  uint16   `flag:"grafana-port" json:"grafana_port"`
	DNSPort                      uint16   `flag:"dns-port" json:"dns_port"`
	APIServerPort                uint16   `json:"api_server_port"`
	NFSData                      string   `flag:"nfs-data" json:"nfs_data"`
	EnableCeph                   bool     `flag:"enable-ceph" json:"enable_ceph"`
	EnableIstio                  bool     `flag:"enable-istio" json:"enable_istio"`
	EnableCNAO                   bool     `flag:"enable-cnao" json:"enable_cnao"`
	EnableNFSCSI                 bool     `flag:"enable-nfs-csi" json:"enable_nfs_csi"`
	EnablePrometheus             bool     `flag:"enable-prometheus" json:"enable_prometheus"`
	EnablePrometheusAlertManager bool     `flag:"enable-prometheus-alertmanager" json:"enable_prometheus_alertmanager"`
	EnableGrafana                bool     `flag:"enable-grafana" json:"enable_grafana"`
	DockerProxy                  string   `flag:"docker-proxy" json:"docker_proxy"`
	GPU                          string   `flag:"gpu" json:"gpu"`
	NvmeDisks                    []string `flag:"nvme" json:"nvme"`
	ScsiDisks                    []string `flag:"scsi" json:"scsi"`
	RunEtcdOnMemory              bool     `flag:"run-etcd-on-memory" json:"run_etcd_on_memory"`
	EtcdCapacity                 string   `flag:"etcd-capacity" json:"etcd_capacity"`
	Hugepages2M                  uint     `flag:"hugepages-2m" json:"hugepages_2m"`
	EnableRealtimeScheduler      bool     `flag:"enable-realtime-scheduler" json:"enable_realtime_scheduler"`
	EnableFIPS                   bool     `flag:"enable-fips" json:"enable_fips"`
	EnablePSA                    bool     `flag:"enable-psa" json:"enable_psa"`
	SingleStack                  bool     `flag:"single-stack" json:"single_stack"`
	EnableAudit                  bool     `flag:"enable-audit" json:"enable_audit"`
	USBDisks                     []string `flag:"usb" json:"usb"`
}

type KubevirtProviderOption func(c *KubevirtProvider)

type FlagConfig struct {
	FlagType        string
	ProviderOptFunc func(interface{}) KubevirtProviderOption
}

var FlagMap = map[string]FlagConfig{
	"nodes": {
		FlagType:        "uint",
		ProviderOptFunc: WithNodes,
	},
	"numa": {
		FlagType:        "uint",
		ProviderOptFunc: WithNuma,
	},
	"memory": {
		FlagType:        "string",
		ProviderOptFunc: WithMemory,
	},
	"cpu": {
		FlagType:        "uint",
		ProviderOptFunc: WithCPU,
	},
	"secondary-nics": {
		FlagType:        "uint",
		ProviderOptFunc: WithSecondaryNics,
	},
	"qemu-args": {
		FlagType:        "string",
		ProviderOptFunc: WithQemuArgs,
	},
	"kernel-args": {
		FlagType:        "string",
		ProviderOptFunc: WithKernelArgs,
	},
	"background": {
		FlagType:        "bool",
		ProviderOptFunc: WithBackground,
	},
	"reverse": {
		FlagType:        "bool",
		ProviderOptFunc: WithReverse,
	},
	"random-ports": {
		FlagType:        "bool",
		ProviderOptFunc: WithRandomPorts,
	},
	"slim": {
		FlagType:        "bool",
		ProviderOptFunc: WithSlim,
	},
	"vnc-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithVNCPort,
	},
	"http-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithHTTPPort,
	},
	"https-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithHTTPSPort,
	},
	"registry-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithRegistryPort,
	},
	"ocp-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithOCPort,
	},
	"k8s-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithK8sPort,
	},
	"ssh-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithSSHPort,
	},
	"prometheus-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithPrometheusPort,
	},
	"grafana-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithGrafanaPort,
	},
	"dns-port": {
		FlagType:        "uint16",
		ProviderOptFunc: WithDNSPort,
	},
	"nfs-data": {
		FlagType:        "string",
		ProviderOptFunc: WithNFSData,
	},
	"enable-ceph": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnableCeph,
	},
	"enable-istio": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnableIstio,
	},
	"enable-cnao": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnableCNAO,
	},
	"enable-nfs-csi": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnableNFSCSI,
	},
	"enable-prometheus": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnablePrometheus,
	},
	"enable-prometheus-alertmanager": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnablePrometheusAlertManager,
	},
	"enable-grafana": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnableGrafana,
	},
	"docker-proxy": {
		FlagType:        "string",
		ProviderOptFunc: WithDockerProxy,
	},
	"gpu": {
		FlagType:        "string",
		ProviderOptFunc: WithGPU,
	},
	"nvme": {
		FlagType:        "[]string",
		ProviderOptFunc: WithNvmeDisks,
	},
	"scsi": {
		FlagType:        "[]string",
		ProviderOptFunc: WithScsiDisks,
	},
	"run-etcd-on-memory": {
		FlagType:        "bool",
		ProviderOptFunc: WithRunEtcdOnMemory,
	},
	"etcd-capacity": {
		FlagType:        "string",
		ProviderOptFunc: WithEtcdCapacity,
	},
	"hugepages-2m": {
		FlagType:        "uint",
		ProviderOptFunc: WithHugepages2M,
	},
	"enable-realtime-scheduler": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnableRealtimeScheduler,
	},
	"enable-fips": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnableFIPS,
	},
	"enable-psa": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnablePSA,
	},
	"single-stack": {
		FlagType:        "bool",
		ProviderOptFunc: WithSingleStack,
	},
	"enable-audit": {
		FlagType:        "bool",
		ProviderOptFunc: WithEnableAudit,
	},
	"usb": {
		FlagType:        "[]string",
		ProviderOptFunc: WithUSBDisks,
	},
}
