package realtime

import utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"

type RealtimeOpt struct {
	sshPort   uint16
	nodeIdx   int
	sshClient utils.SSHClient
}

func NewRealtimeOpt(sc utils.SSHClient, sshPort uint16, nodeIdx int) *RealtimeOpt {
	return &RealtimeOpt{
		sshPort:   sshPort,
		nodeIdx:   nodeIdx,
		sshClient: sc,
	}
}

func (o *RealtimeOpt) Exec() error {
	cmds := []string{
		"echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf",
		"sysctl --system",
	}

	for _, cmd := range cmds {
		if _, err := o.sshClient.JumpSSH(o.sshPort, o.nodeIdx, cmd, true, true); err != nil {
			return err
		}
	}
	return nil
}
