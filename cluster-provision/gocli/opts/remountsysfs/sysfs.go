package remountsysfs

import "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"

type RemountSysFSOpt struct {
	sshClient libssh.Client
}

func NewRemountSysFSOpt(sshClient libssh.Client) *RemountSysFSOpt {
	return &RemountSysFSOpt{
		sshClient: sshClient,
	}
}

func (r *RemountSysFSOpt) Exec() error {
	cmds := []string{
		"mount -o remount,rw /sys",
		"ls -la -Z /dev/vfio",
		"chmod 0666 /dev/vfio/vfio",
	}

	for _, cmd := range cmds {
		if err := r.sshClient.Command(cmd); err != nil {
			return err
		}
	}
	return nil
}
