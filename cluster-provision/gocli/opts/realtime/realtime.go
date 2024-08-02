package realtime

import (
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type realtimeOpt struct {
	sshClient libssh.Client
}

func NewRealtimeOpt(sc libssh.Client) *realtimeOpt {
	return &realtimeOpt{
		sshClient: sc,
	}
}

func (o *realtimeOpt) Exec() error {
	cmds := []string{
		"echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf",
		"sysctl --system",
	}

	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}
	return nil
}
