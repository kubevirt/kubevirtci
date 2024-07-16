package realtime

import (
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type RealtimeOpt struct {
	sshClient libssh.Client
}

func NewRealtimeOpt(sc libssh.Client) *RealtimeOpt {
	return &RealtimeOpt{
		sshClient: sc,
	}
}

func (o *RealtimeOpt) Exec() error {
	cmds := []string{
		"echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf",
		"sysctl --system",
	}

	for _, cmd := range cmds {
		if _, err := o.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}
	return nil
}
