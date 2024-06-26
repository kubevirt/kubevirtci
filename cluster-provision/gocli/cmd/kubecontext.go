package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/providers"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func NewSetContextCommand() *cobra.Command {
	ctxCmd := &cobra.Command{
		Use:   "set-context",
		Short: "adds the cluster created by cluster up to your kube context for your own use",
		RunE:  setKubeContext,
		Args:  cobra.ExactArgs(1),
	}

	return ctxCmd

}

func setKubeContext(cmd *cobra.Command, args []string) error {
	prefix := args[0]
	kp, err := providers.NewFromRunning(prefix)
	if err != nil {
		return err
	}

	err = kp.SSHClient.CopyRemoteFile(kp.SSHPort, "/etc/kubernetes/admin.conf", ".tempkubeconfig")
	if err != nil {
		return err
	}

	conf, err := k8s.InitConfig(".tempkubeconfig", kp.APIServerPort)
	if err != nil {
		return err
	}

	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["kubevirtci"] = &clientcmdapi.Cluster{
		Server:                   conf.Host,
		CertificateAuthorityData: []byte{},
		InsecureSkipTLSVerify:    true,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contexts["kubevirtci-context"] = &clientcmdapi.Context{
		Cluster:  "kubevirtci",
		AuthInfo: "kubevirtci-admin",
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["kubevirtci-admin"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: conf.CertData,
		ClientKeyData:         conf.KeyData,
	}
	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "kubevirtci-context",
		AuthInfos:      authinfos,
	}
	kubeConfigFile, err := os.Create("/tmp/.kubeconfig")
	if err != nil {
		return err
	}

	err = clientcmd.WriteToFile(clientConfig, kubeConfigFile.Name())
	if err != nil {
		return err
	}
	// todo: account for an existing kubeconfig
	setenv := exec.Command("sh", "-c", "export KUBECONFIG=$KUBECONFIG:/tmp/.kubeconfig >> /etc/environment")
	setenv.Stdout = os.Stdout
	setenv.Stderr = os.Stderr
	err = setenv.Run()
	if err != nil {
		return err
	}

	return nil
}
