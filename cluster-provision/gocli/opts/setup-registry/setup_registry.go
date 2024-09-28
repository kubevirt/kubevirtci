package setupregistry

import "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"

type SetupRegistryOpt struct {
	sshClient  libssh.Client
	registryIP string
}

func NewSetupRegistry(sshClient libssh.Client, registryIP string) *SetupRegistryOpt {
	return &SetupRegistryOpt{
		sshClient:  sshClient,
		registryIP: registryIP,
	}
}

func (sr *SetupRegistryOpt) Exec() error {
	cmds := []string{
		"echo " + sr.registryIP + "\tregistry | tee -a /etc/hosts",
	}
	for _, cmd := range cmds {
		if err := sr.sshClient.Command(cmd); err != nil {
			return err
		}
	}
	return nil
}
