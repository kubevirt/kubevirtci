package cmd

import (
	"context"
	"os"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	sshutils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

func NewSetContextCommand() *cobra.Command {
	ctxCmd := &cobra.Command{
		Use:   "set-context",
		Short: "adds the cluster created by cluster up to your kube context for your own use",
		RunE:  setKubeContext,
	}

	return ctxCmd

}

// still use the old method of reading the port but code the new one too
func setKubeContext(cmd *cobra.Command, args []string) error {
	// kp, err := providers.NewFromRunning(prefix)
	// if err != nil {
	// 	return err
	// }

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	containers, err := docker.GetPrefixedContainers(cli, prefix+"-dnsmasq")
	if err != nil {
		return err
	}

	container, err := cli.ContainerInspect(context.Background(), containers[0].ID)
	if err != nil {
		return err
	}

	apiServerPort, err := utils.GetPublicPort(utils.PortAPI, container.NetworkSettings.Ports)
	if err != nil {
		return err
	}

	sshPort, err := utils.GetPublicPort(utils.PortSSH, container.NetworkSettings.Ports)
	if err != nil {
		return err
	}

	err = sshutils.CopyRemoteFile(sshPort, "/etc/kubernetes/admin.conf", ".tempkubeconfig")
	if err != nil {
		return err
	}

	// err = utils.CopyRemoteFile(kp.SSHPort, "/etc/kubernetes/admin.conf", ".tempkubeconfig")
	conf, err := k8s.InitConfig(".tempkubeconfig", apiServerPort)
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

	err = os.Setenv("KUBECONFIG", "/tmp/.kubeconfig")
	if err != nil {
		return err
	}

	return nil
}
