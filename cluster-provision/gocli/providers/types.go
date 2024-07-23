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
	Hugepages2M                  uint     `flag:"hugepages-2m" json:"hugepages_2m"`
	Hugepages1G                  uint     `flag:"hugepages-1g" json:"hugepages_1g"`
	EnableRealtimeScheduler      bool     `flag:"enable-realtime-scheduler" json:"enable_realtime_scheduler"`
	EnableFIPS                   bool     `flag:"enable-fips" json:"enable_fips"`
	EnablePSA                    bool     `flag:"enable-psa" json:"enable_psa"`
	SingleStack                  bool     `flag:"single-stack" json:"single_stack"`
	EnableAudit                  bool     `flag:"enable-audit" json:"enable_audit"`
}

type KubevirtProviderOption func(c *KubevirtProvider)
