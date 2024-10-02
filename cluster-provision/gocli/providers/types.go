package providers

import (
	"github.com/docker/docker/client"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

type KubevirtProvider struct {
	Client  k8s.K8sDynamicClient `json:"-"`
	Docker  *client.Client       `json:"-"`
	DNSMasq string               `json:"dnsmasq"`

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
	RandomPorts                  bool     `flag:"random-ports" json:"random_ports"`
	Prefix                       string   `flag:"prefix" json:"prefix"`
	Slim                         bool     `flag:"slim" json:"slim"`
	VNCPort                      uint     `flag:"vnc-port" json:"vnc_port"`
	HTTPPort                     uint     `flag:"http-port" json:"http_port"`
	HTTPSPort                    uint     `flag:"https-port" json:"https_port"`
	RegistryPort                 uint     `flag:"registry-port" json:"registry_port"`
	OCPort                       uint     `flag:"ocp-port" json:"ocp_port"`
	K8sPort                      uint     `flag:"k8s-port" json:"k8s_port"`
	SSHPort                      uint     `flag:"ssh-port" json:"ssh_port"`
	PrometheusPort               uint     `flag:"prometheus-port" json:"prometheus_port"`
	GrafanaPort                  uint     `flag:"grafana-port" json:"grafana_port"`
	DNSPort                      uint     `flag:"dns-port" json:"dns_port"`
	APIServerPort                uint     `json:"api_server_port"`
	NFSData                      string   `flag:"nfs-data" json:"nfs_data"`
	EnableCeph                   bool     `flag:"enable-ceph" json:"enable_ceph"`
	EnableIstio                  bool     `flag:"enable-istio" json:"enable_istio"`
	EnableCNAO                   bool     `flag:"enable-cnao" json:"enable_cnao"`
	SkipCnaoCR                   bool     `flag:"skip-cnao-cr" json:"skip_cnao_cr"`
	EnableNFSCSI                 bool     `flag:"enable-nfs-csi" json:"enable_nfs_csi"`
	EnablePrometheus             bool     `flag:"enable-prometheus" json:"enable_prometheus"`
	EnablePrometheusAlertManager bool     `flag:"enable-prometheus-alertmanager" json:"enable_prometheus_alertmanager"`
	EnableGrafana                bool     `flag:"enable-grafana" json:"enable_grafana"`
	EnableMultus                 bool     `flag:"deploy-multus" json:"deploy_multus"`
	DockerProxy                  string   `flag:"docker-proxy" json:"docker_proxy"`
	AAQ                          bool     `flag:"deploy-aaq" json:"deploy_aaq"`
	AAQVersion                   string   `flag:"aaq-version" json:"aaq_version"`
	CDI                          bool     `flag:"deploy-cdi" json:"deploy_cdi"`
	CDIVersion                   string   `flag:"cdi-version" json:"cdi_version"`
	GPU                          string   `flag:"gpu" json:"gpu"`
	KSM                          bool     `flag:"enable-ksm" json:"enable_ksm"`
	KSMPages                     uint     `flag:"ksm-page-count" json:"ksm_page_count"`
	KSMInterval                  uint     `flag:"ksm-scan-interval" json:"ksm_scan_interval"`
	Swap                         bool     `flag:"enable-swap" json:"enable_swap"`
	Swapsize                     uint     `flag:"swap-size" json:"swap_size"`
	UnlimitedSwap                bool     `flag:"unlimited-swap" json:"unlimited_swap"`
	Swapiness                    uint     `flag:"swapiness" json:"swapiness"`
	NvmeDisks                    []string `flag:"nvme" json:"nvme"`
	ScsiDisks                    []string `flag:"scsi" json:"scsi"`
	USBDisks                     []string `flag:"usb" json:"usb"`
	AdditionalKernelArgs         []string `flag:"additional-persistent-kernel-arguments" json:"additional-persistent-kernel-arguments"`
	Phases                       string   `flag:"phases" json:"phases"`
	RunEtcdOnMemory              bool     `flag:"run-etcd-on-memory" json:"run_etcd_on_memory"`
	EtcdCapacity                 string   `flag:"etcd-capacity" json:"etcd_capacity"`
	NoEtcdFsync                  bool     `flag:"no-etcd-fsync" json:"no_etcd_fsync"`
	Hugepages2M                  uint     `flag:"hugepages-2m" json:"hugepages_2m"`
	Hugepages1G                  uint     `flag:"hugepages-1g" json:"hugepages_1g"`
	EnableRealtimeScheduler      bool     `flag:"enable-realtime-scheduler" json:"enable_realtime_scheduler"`
	EnableFIPS                   bool     `flag:"enable-fips" json:"enable_fips"`
	EnablePSA                    bool     `flag:"enable-psa" json:"enable_psa"`
	SingleStack                  bool     `flag:"single-stack" json:"single_stack"`
	EnableAudit                  bool     `flag:"enable-audit" json:"enable_audit"`
}

type KubevirtProviderOption func(c *KubevirtProvider)

type FlagConfig struct {
	FlagType        string
	ProviderOptFunc func(interface{}) KubevirtProviderOption
}

var ProvisionFlagMap = map[string]FlagConfig{
	"memory": {
		FlagType:        "string",
		ProviderOptFunc: WithMemory,
	},
	"cpu": {
		FlagType:        "uint",
		ProviderOptFunc: WithCPU,
	},
	"slim": {
		FlagType:        "bool",
		ProviderOptFunc: WithSlim,
	},
	"random-ports": {
		FlagType:        "bool",
		ProviderOptFunc: WithRandomPorts,
	},
	"phases": {
		FlagType:        "string",
		ProviderOptFunc: WithPhases,
	},
	"additional-persistent-kernel-arguments": {
		FlagType:        "[]string",
		ProviderOptFunc: WithAdditionalKernelArgs,
	},
	"vnc-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithVNCPort,
	},
	"ssh-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithSSHPort,
	},
	"qemu-args": {
		FlagType:        "string",
		ProviderOptFunc: WithQemuArgs,
	},
}

var RunFlagMap = map[string]FlagConfig{
	"nodes": {
		FlagType:        "uint",
		ProviderOptFunc: WithNodes,
	},
	"prefix": {
		FlagType:        "string",
		ProviderOptFunc: WithPrefix,
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
	"random-ports": {
		FlagType:        "bool",
		ProviderOptFunc: WithRandomPorts,
	},
	"slim": {
		FlagType:        "bool",
		ProviderOptFunc: WithSlim,
	},
	"vnc-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithVNCPort,
	},
	"http-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithHTTPPort,
	},
	"https-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithHTTPSPort,
	},
	"registry-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithRegistryPort,
	},
	"ocp-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithOCPort,
	},
	"k8s-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithK8sPort,
	},
	"ssh-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithSSHPort,
	},
	"prometheus-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithPrometheusPort,
	},
	"grafana-port": {
		FlagType:        "uint",
		ProviderOptFunc: WithGrafanaPort,
	},
	"dns-port": {
		FlagType:        "uint",
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
	"deploy-multus": {
		FlagType:        "bool",
		ProviderOptFunc: WithMultus,
	},
	"deploy-aaq": {
		FlagType:        "bool",
		ProviderOptFunc: WithAAQ,
	},
	"deploy-cdi": {
		FlagType:        "bool",
		ProviderOptFunc: WithCDI,
	},
	"enable-ksm": {
		FlagType:        "bool",
		ProviderOptFunc: WithKSM,
	},
	"ksm-page-count": {
		FlagType:        "uint",
		ProviderOptFunc: WithKSMPages,
	},
	"ksm-scan-interval": {
		FlagType:        "uint",
		ProviderOptFunc: WithKSMInterval,
	},
	"enable-swap": {
		FlagType:        "bool",
		ProviderOptFunc: WithSwap,
	},
	"unlimited-swap": {
		FlagType:        "bool",
		ProviderOptFunc: WithUnlimitedSwap,
	},
	"swap-size": {
		FlagType:        "uint",
		ProviderOptFunc: WithSwapSize,
	},
	"swapiness": {
		FlagType:        "uint",
		ProviderOptFunc: WithSwapiness,
	},
	"cdi-version": {
		FlagType:        "string",
		ProviderOptFunc: WithCDIVersion,
	},
	"aaq-version": {
		FlagType:        "string",
		ProviderOptFunc: WithAAQVersion,
	},
	"skip-cnao-cr": {
		FlagType:        "bool",
		ProviderOptFunc: WithSkipCnaoCR,
	},
	"no-etcd-fsync": {
		FlagType:        "bool",
		ProviderOptFunc: WithNoEtcdFsync,
	},
}
