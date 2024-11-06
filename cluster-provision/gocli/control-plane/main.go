package controlplane

import (
	"path"

	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri/podman"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

const (
	etcdImage         = "etcd:3.5.10-0"
	apiServer         = "kube-apiserver"
	controllerManager = "kube-controller-manager"
	scheduler         = "kube-scheduler"
	registry          = "registry.k8s.io"
	defaultPkiPath    = "/etc/kubevirtci/pki"
)

type ControlPlaneRunner struct {
	dnsmasqID        string
	containerRuntime cri.ContainerClient
	Client           k8s.K8sDynamicClient
	k8sVersion       string
	apiServerPort    uint
}

type Phase interface {
	Run() error
}

func NewControlPlaneRunner(dnsmasqID string, k8sVersion string, apiServerPort uint) *ControlPlaneRunner {
	var containerRuntime cri.ContainerClient
	if podman.IsAvailable() {
		containerRuntime = podman.NewPodman()

	}
	if true {
		containerRuntime = docker.NewDockerClient()
	}

	return &ControlPlaneRunner{
		containerRuntime: containerRuntime,
		dnsmasqID:        dnsmasqID,
		k8sVersion:       k8sVersion,
		apiServerPort:    apiServerPort,
	}
}

func (cp *ControlPlaneRunner) Start() (*rest.Config, error) {
	if err := NewCertsPhase(defaultPkiPath).Run(); err != nil {
		return nil, err
	}

	if err := NewRunETCDPhase(cp.dnsmasqID, cp.containerRuntime, defaultPkiPath, cp.k8sVersion).Run(); err != nil {
		return nil, err
	}

	if err := NewKubeConfigPhase(defaultPkiPath).Run(); err != nil {
		return nil, err
	}

	if err := NewRunControlPlaneComponentsPhase(cp.dnsmasqID, cp.containerRuntime, defaultPkiPath, cp.k8sVersion).Run(); err != nil {
		return nil, err
	}

	config, err := k8s.NewConfig(path.Join(defaultPkiPath, "admin/.kubeconfig"), uint16(cp.apiServerPort))
	if err != nil {
		return nil, err
	}

	k8sClient, err := k8s.NewDynamicClient(config)
	if err != nil {
		return nil, err
	}

	if err := NewCreateBootstrappersRBACPhase(k8sClient).Run(); err != nil {
		return nil, err
	}

	if err := NewBootstrapAuthResourcesPhase(k8sClient, defaultPkiPath).Run(); err != nil {
		return nil, err
	}

	if err := NewKubeProxyPhase(k8sClient).Run(); err != nil {
		return nil, err
	}

	if err := NewCNIPhase(k8sClient, false).Run(); err != nil {
		return nil, err
	}

	if err := NewCoreDNSPhase(k8sClient).Run(); err != nil {
		return nil, err
	}

	return config, nil
}

func RestConfigToKubeconfig(restConfig *rest.Config) *clientcmdapi.Config {
	kubeconfig := clientcmdapi.NewConfig()
	kubeconfig.Clusters["default"] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		CertificateAuthorityData: restConfig.CAData,
		InsecureSkipTLSVerify:    restConfig.Insecure,
	}

	kubeconfig.AuthInfos["default"] = &clientcmdapi.AuthInfo{
		Token:                 restConfig.BearerToken,
		ClientCertificateData: restConfig.CertData,
		ClientKeyData:         restConfig.KeyData,
		Username:              restConfig.Username,
		Password:              restConfig.Password,
	}

	kubeconfig.Contexts["default"] = &clientcmdapi.Context{
		Cluster:  "default",
		AuthInfo: "default",
	}

	kubeconfig.CurrentContext = "default"
	return kubeconfig
}
