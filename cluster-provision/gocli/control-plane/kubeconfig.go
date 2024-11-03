package controlplane

import (
	"os"
	"fmt"
	"path"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type KubeConfigPhase struct {
	pkiPath string
}

var defaultComponents = map[string]string{
	"admin":                   "admin",
	"kube-scheduler":          "system:kube-scheduler",
	"kube-controller-manager": "system:kube-controller-manager"}

func NewKubeConfigPhase(pkiPath string) *KubeConfigPhase {
	return &KubeConfigPhase{
		pkiPath: pkiPath,
	}
}

func (p *KubeConfigPhase) Run() error {
	ca, err := os.ReadFile(path.Join(p.pkiPath, "ca.crt"))
	if err != nil {
		return err
	}

	for component, userName := range defaultComponents {
		clientCert, err := os.ReadFile(path.Join(p.pkiPath, component+".crt"))
		if err != nil {
			return err
		}

		key, err := os.ReadFile(path.Join(p.pkiPath, component+".pem"))
		if err != nil {
			return err
		}

		kubeconfig := buildKubeConfigFromCerts(ca, clientCert, key, "https://127.0.0.1:6443", userName) // todo: handle this better
		err = clientcmd.WriteToFile(kubeconfig, p.pkiPath+"/"+component+"/.kubeconfig")
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	
	}
	return nil
}

func buildKubeConfigFromCerts(ca, clientCert, clientKey []byte, server, user string) clientcmdapi.Config {
	clusters := map[string]*clientcmdapi.Cluster{
		"kubernetes": {
			Server:                   server,
			CertificateAuthorityData: ca,
		},
	}
	contexts := map[string]*clientcmdapi.Context{
		"default": {
			Cluster:  "kubernetes",
			AuthInfo: user,
		},
	}
	authInfos := map[string]*clientcmdapi.AuthInfo{
		user: {
			ClientCertificateData: clientCert,
			ClientKeyData:         clientKey,
		},
	}

	return clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default",
		AuthInfos:      authInfos,
	}
}
