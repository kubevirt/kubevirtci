package kindbase

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/docker/docker/client"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"

	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
	dockercri "kubevirt.io/kubevirtci/cluster-provision/gocli/cri/docker"
	podmancri "kubevirt.io/kubevirtci/cluster-provision/gocli/cri/podman"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/network"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/registryproxy"
	setupregistry "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/setup-registry"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
	kind "sigs.k8s.io/kind/pkg/cluster"
)

//go:embed manifests/*
var f embed.FS

type KindProvider interface {
	Start(ctx context.Context, cancel context.CancelFunc) error
	Delete() error
}

type KindBaseProvider struct {
	Client   k8s.K8sDynamicClient
	CRI      cri.ContainerClient
	Provider *kind.Provider
	Image    string
	Cluster  string

	*KindConfig
}
type KindConfig struct {
	Nodes           int
	RegistryPort    string
	Version         string
	RunEtcdOnMemory bool
	IpFamily        string
	WithCPUManager  bool
	RegistryProxy   string
	WithExtraMounts bool
	WithVfio        bool
}

const (
	kind128Image      = "kindest/node:v1.28.0@sha256:b7a4cad12c197af3ba43202d3efe03246b3f0793f162afb40a33c923952d5b31"
	cniArchieFilename = "cni-archive.tar.gz"
	registryImage     = "quay.io/kubevirtci/library-registry:2.7.1"
)

func NewKindBaseProvider(kindConfig *KindConfig) (*KindBaseProvider, error) {
	var (
		cri cri.ContainerClient
		k   *kind.Provider
	)

	runtime, err := DetectContainerRuntime()
	if err != nil {
		return nil, err
	}

	switch runtime {
	case "docker":
		logrus.Info("Using Docker as container runtime")
		cri = dockercri.NewDockerClient()
		k = kind.NewProvider(kind.ProviderWithDocker())
	case "podman":
		logrus.Info("Using Podman as container runtime")
		cri = podmancri.NewPodman()
		k = kind.NewProvider(kind.ProviderWithPodman())
	}

	kp := &KindBaseProvider{
		Image:      kind128Image,
		CRI:        cri,
		Provider:   k,
		KindConfig: kindConfig,
	}
	cluster, err := kp.PrepareClusterYaml(kindConfig.WithExtraMounts, kindConfig.WithVfio)
	if err != nil {
		return nil, err
	}

	kp.Cluster = cluster
	return kp, nil
}

func DetectContainerRuntime() (string, error) {
	if podmancri.IsAvailable() {
		return "podman", nil
	}
	if dockercri.IsAvailable() {
		return "docker", nil
	}
	return "", fmt.Errorf("No valid container runtime found")
}

func (k *KindBaseProvider) Start(ctx context.Context, cancel context.CancelFunc) error {
	err := k.Provider.Create(k.Version, kind.CreateWithRawConfig([]byte(k.Cluster)), kind.CreateWithNodeImage(k.Image))
	if err != nil {
		return err
	}
	logrus.Infof("Kind %s base cluster started\n", k.Version)

	kubeconf, err := k.Provider.KubeConfig(k.Version, true)
	if err != nil {
		return err
	}

	jsonData, err := yaml.YAMLToJSON([]byte(kubeconf))
	if err != nil {
		return err
	}
	config := &rest.Config{}
	err = json.Unmarshal(jsonData, config)
	if err != nil {
		return err
	}

	k8sClient, err := k8s.NewDynamicClient(config)
	if err != nil {
		return err
	}
	k.Client = k8sClient
	nodes, err := k.Provider.ListNodes(k.Version)
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	err = k.downloadCNI()
	if err != nil {
		return nil
	}

	_, registryIP, err := k.runRegistry(k.RegistryPort)
	if err != nil {
		return err
	}

	var sshClient libssh.Client
	for _, node := range nodes {
		switch k.CRI.(type) {
		case *dockercri.DockerClient:
			sshClient = docker.NewAdapter(cli, node.String())
		case *podmancri.Podman:
			sshClient = podmancri.NewPodmanSSHClient(node.String())
		}

		if err := k.setupCNI(sshClient); err != nil {
			return err
		}

		sr := setupregistry.NewSetupRegistry(sshClient, registryIP)
		if err = sr.Exec(); err != nil {
			return err
		}

		n := network.NewNetworkOpt(sshClient)
		if err = n.Exec(); err != nil {
			return err
		}

		if k.RegistryProxy != "" {
			rp := registryproxy.NewRegistryProxyOpt(sshClient, k.RegistryProxy)
			if err = rp.Exec(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (k *KindBaseProvider) Delete() error {
	if err := k.Provider.Delete(k.Version, ""); err != nil {
		return err
	}
	if err := k.deleteRegistry(); err != nil {
		return err
	}
	return nil
}

func (k *KindBaseProvider) PrepareClusterYaml(withExtraMounts, withVfio bool) (string, error) {
	cluster, err := f.ReadFile("manifests/kind.yaml")
	if err != nil {
		return "", err
	}

	wp, err := f.ReadFile("manifests/worker-patch.yaml")
	if err != nil {
		return "", err
	}

	cpump, err := f.ReadFile("manifests/cpu-manager-patch.yaml")
	if err != nil {
		return "", err
	}

	ipf, err := f.ReadFile("manifests/ip-family.yaml")
	if err != nil {
		return "", err
	}

	if withExtraMounts {
		aud, err := f.ReadFile("manifests/audit.yaml")
		if err != nil {
			return "", err
		}
		cluster = append(cluster, aud...)
		cluster = append(cluster, []byte("\n")...)
	}

	if withVfio {
		vfio, err := f.ReadFile("manifests/vfio.yaml")
		if err != nil {
			return "", err
		}
		cluster = append(cluster, vfio...)
		cluster = append(cluster, []byte("\n")...)
	}

	for i := 0; i < k.Nodes; i++ {
		cluster = append(cluster, wp...)
		cluster = append(cluster, []byte("\n")...)
		if k.WithCPUManager {
			cluster = append(cluster, cpump...)
			cluster = append(cluster, []byte("\n")...)
		}
	}

	if k.IpFamily != "" {
		cluster = append(cluster, []byte(string(ipf)+k.IpFamily)...)
	}
	return string(cluster), nil
}

func (k *KindBaseProvider) setupCNI(sshClient libssh.Client) error {
	file, err := os.Open(cniArchieFilename)
	if err != nil {
		return err
	}

	err = sshClient.SCP("/opt/cni/bin", file)
	if err != nil {
		return err
	}
	return nil
}

func (k *KindBaseProvider) deleteRegistry() error {
	return k.CRI.Remove(k.Version + "-registry")
}

func (k *KindBaseProvider) runRegistry(hostPort string) (string, string, error) {
	registryID, err := k.CRI.Create(registryImage, &cri.CreateOpts{
		Name:          k.Version + "-registry",
		Privileged:    true,
		Network:       "kind",
		RestartPolicy: "always",
		Ports: map[string]string{
			"5000": hostPort,
		},
	})
	if err != nil {
		return "", "", err
	}

	if err := k.CRI.Start(registryID); err != nil {
		return "", "", err
	}

	ip, err := k.CRI.Inspect(registryID, "{{.NetworkSettings.Networks.kind.IPAddress}}")
	if err != nil {
		return "", "", err
	}

	return registryID, strings.TrimSuffix(string(ip), "\n"), nil
}

func (k *KindBaseProvider) downloadCNI() error {
	out, err := os.Create(cniArchieFilename)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get("https://github.com/containernetworking/plugins/releases/download/v0.8.5/cni-plugins-linux-" + runtime.GOARCH + "-v0.8.5.tgz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	logrus.Info("Downloaded cni archive")
	return nil
}
