package realtime

import utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"

type RealtimeOpt struct {
	sshPort uint16
	nodeIdx int
}

func NewRealtimeOpt(sshPort uint16, nodeIdx int) *RealtimeOpt {
	return &RealtimeOpt{
		sshPort: sshPort,
		nodeIdx: nodeIdx,
	}
}

func (o *RealtimeOpt) Exec() error {
	cmds := []string{
		"echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf",
		"sysctl --system",
	}

	for _, cmd := range cmds {
		if _, err := utils.JumpSSH(o.sshPort, 1, cmd, true, true); err != nil {
			return err
		}
	}
	return nil
}
