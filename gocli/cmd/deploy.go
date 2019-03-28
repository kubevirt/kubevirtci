package cmd

import (
	"fmt"
	"github.com/ghodss/yaml"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

func NewDeployCommand() *cobra.Command {

	run := &cobra.Command{
		Use:   "deploy",
		Short: "deploy deploys a cluster as a Pod on another cluster",
		RunE:  runDeploy,
		Args:  cobra.ExactArgs(1),
	}
	run.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	run.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	run.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	run.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	run.Flags().BoolP("reverse", "r", false, "revert node startup order")
	run.Flags().String("nfs-volume-claim", "", "PVC to use for NFS data")
	run.Flags().String("log-to-dir", "", "enables aggregated cluster logging to the folder")
	run.Flags().Bool("enable-ceph", false, "enables dynamic storage provisioning using Ceph")
	return run
}

func runDeploy(cmd *cobra.Command, args []string) (err error) {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	nodes, err := cmd.Flags().GetUint("nodes")
	if err != nil {
		return err
	}

	_, err = cmd.Flags().GetString("memory")
	if err != nil {
		return err
	}

	_, err = cmd.Flags().GetBool("reverse")
	if err != nil {
		return err
	}

	_, err = cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}

	_, err = cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	nfsClaim, err := cmd.Flags().GetString("nfs-volume-claim")
	if err != nil {
		return err
	}

	_, err = cmd.Flags().GetString("log-to-dir")
	if err != nil {
		return err
	}

	_, err = cmd.Flags().GetBool("enable-ceph")
	if err != nil {
		return err
	}

	pod := v1.Pod{
		TypeMeta: v12.TypeMeta{
			Kind: "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name: prefix,
		},
		Spec: v1.PodSpec{
			HostAliases: []v1.HostAlias{
				{Hostnames: []string{"nfs", "registry", "ceph"}, IP: "192.168.66.2"},
			},
			Volumes: []v1.Volume{
				{
					Name: "registry",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	cluster := args[0]

	// dnsmasq container
	t := true
	dnsContainer := v1.Container{
		Name:    "dnsmasq",
		Image:   cluster,
		Command: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
		Env: []v1.EnvVar{
			{Name: "NUM_NODES", Value: fmt.Sprintf("%v", nodes)},
			{Name: "NEX_HOST", Value: "node01"},
		},
		ImagePullPolicy: "IfNotPresent",
		SecurityContext: &v1.SecurityContext{
			Privileged: &t,

		},
		Ports: []v1.ContainerPort{
			{Name: PORT_NAME_K8S, ContainerPort: PORT_K8S},
			{Name: PORT_NAME_SSH, ContainerPort: PORT_SSH},
			{Name: PORT_NAME_VNC, ContainerPort: PORT_VNC},
			{Name: PORT_NAME_OCP, ContainerPort: PORT_OCP},
		},
	}

	pod.Spec.Containers = append(pod.Spec.Containers, dnsContainer)

	// registry
	registryContainer := v1.Container{
		Name:  "registry",
		Image: DockerRegistryImage,
		Ports: []v1.ContainerPort{
			{Name: PORT_NAME_REGISTRY, ContainerPort: PORT_REGISTRY},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name: "registry",
				MountPath: "/var/lib/registry",
			},
		},
	}

	pod.Spec.Containers = append(pod.Spec.Containers, registryContainer)


	// nfs server
	if nfsClaim != "" {
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "nfsClaim",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: nfsClaim,
				},
			},
		})
		nfsContainer := v1.Container{
			Name:  "nfsClaim",
			Image: NFSGaneshaImage,
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "nfsClaim",
					MountPath: "/data/nfsClaim",
				},
			},
		}
		pod.Spec.Containers = append(pod.Spec.Containers, nfsContainer)
	}

	data, err := yaml.Marshal(pod)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
