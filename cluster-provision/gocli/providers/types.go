package providers

import (
	"github.com/docker/docker/client"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

type KubevirtProvider struct {
	IsRunning bool
	Client    *k8s.K8sDynamicClient
	Docker    *client.Client
	DNSMasq   string

	Version string
	Image   string
	Nodes   uint   `flag:"nodes" short:"n"`
	Numa    uint   `flag:"numa" short:"u"`
	Memory  string `flag:"memory" short:"m"`
	CPU     uint   `flag:"cpu" short:"c"`

	SecondaryNics                uint   `flag:"secondary-nics"`
	QemuArgs                     string `flag:"qemu-args"`
	KernelArgs                   string `flag:"kernel-args"`
	Background                   bool   `flag:"background" short:"b"`
	Reverse                      bool   `flag:"reverse" short:"r"`
	RandomPorts                  bool   `flag:"random-ports"`
	Slim                         bool   `flag:"slim"`
	VNCPort                      uint16 `flag:"vnc-port"`
	HTTPPort                     uint16 `flag:"http-port"`
	HTTPSPort                    uint16 `flag:"https-port"`
	RegistryPort                 uint16 `flag:"registry-port"`
	OCPort                       uint16 `flag:"ocp-port"`
	K8sPort                      uint16 `flag:"k8s-port"`
	SSHPort                      uint16 `flag:"ssh-port"`
	PrometheusPort               uint16 `flag:"prometheus-port"`
	GrafanaPort                  uint16 `flag:"grafana-port"`
	DNSPort                      uint16 `flag:"dns-port"`
	APIServerPort                uint16
	NFSData                      string   `flag:"nfs-data"`
	EnableCeph                   bool     `flag:"enable-ceph"`
	EnableIstio                  bool     `flag:"enable-istio"`
	EnableCNAO                   bool     `flag:"enable-cnao"`
	EnableNFSCSI                 bool     `flag:"enable-nfs-csi"`
	EnablePrometheus             bool     `flag:"enable-prometheus"`
	EnablePrometheusAlertManager bool     `flag:"enable-prometheus-alertmanager"`
	EnableGrafana                bool     `flag:"enable-grafana"`
	DockerProxy                  string   `flag:"docker-proxy"`
	GPU                          string   `flag:"gpu"`
	NvmeDisks                    []string `flag:"nvme"`
	ScsiDisks                    []string `flag:"scsi"`
	RunEtcdOnMemory              bool     `flag:"run-etcd-on-memory"`
	EtcdCapacity                 string   `flag:"etcd-capacity"`
	Hugepages2M                  uint     `flag:"hugepages-2m"`
	EnableRealtimeScheduler      bool     `flag:"enable-realtime-scheduler"`
	EnableFIPS                   bool     `flag:"enable-fips"`
	EnablePSA                    bool     `flag:"enable-psa"`
	SingleStack                  bool     `flag:"single-stack"`
	EnableAudit                  bool     `flag:"enable-audit"`
	USBDisks                     []string `flag:"usb"`
}

type KubevirtProviderOption func(*KubevirtProvider)
