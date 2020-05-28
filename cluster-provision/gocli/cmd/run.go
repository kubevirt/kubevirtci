package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/okd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

const proxySettings = `
mkdir -p /etc/systemd/system/docker.service.d/

cat <<EOT  >/etc/systemd/system/docker.service.d/proxy.conf
[Service]
Environment="HTTP_PROXY={{.Proxy}}"
Environment="HTTPS_PROXY={{.Proxy}}"
Environment="NO_PROXY=localhost,127.0.0.1"
EOT

systemctl daemon-reload
systemctl restart docker
EOF
`

type dockerSetting struct {
	Proxy string
}

// NewRunCommand returns command that runs given cluster
func NewRunCommand() *cobra.Command {

	run := &cobra.Command{
		Use:   "run",
		Short: "run starts a given cluster",
		RunE:  run,
		Args:  cobra.ExactArgs(1),
	}
	run.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	run.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	run.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	run.Flags().UintP("secondary-nics", "", 0, "number of secondary nics to add")
	run.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	run.Flags().BoolP("background", "b", false, "go to background after nodes are up")
	run.Flags().BoolP("reverse", "r", false, "revert node startup order")
	run.Flags().Bool("random-ports", true, "expose all ports on random localhost ports")
	run.Flags().String("registry-volume", "", "cache docker registry content in the specified volume")
	run.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	run.Flags().Uint("registry-port", 0, "port on localhost for the docker registry")
	run.Flags().Uint("ocp-port", 0, "port on localhost for the ocp cluster")
	run.Flags().Uint("k8s-port", 0, "port on localhost for the k8s cluster")
	run.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	run.Flags().String("nfs-data", "", "path to data which should be exposed via nfs to the nodes")
	run.Flags().String("log-to-dir", "", "enables aggregated cluster logging to the folder")
	run.Flags().Bool("enable-ceph", false, "enables dynamic storage provisioning using Ceph")
	run.Flags().String("docker-proxy", "", "sets network proxy for docker daemon")
	run.Flags().String("container-registry", "docker.io", "the registry to pull cluster container from")
	run.Flags().Bool("enable-ovsdpdk", false, "enable host preparation for ovsdpdk")

	run.AddCommand(
		okd.NewRunCommand(),
	)
	return run
}

func run(cmd *cobra.Command, args []string) (err error) {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	nodes, err := cmd.Flags().GetUint("nodes")
	if err != nil {
		return err
	}

	memory, err := cmd.Flags().GetString("memory")
	if err != nil {
		return err
	}

	reverse, err := cmd.Flags().GetBool("reverse")
	if err != nil {
		return err
	}

	randomPorts, err := cmd.Flags().GetBool("random-ports")
	if err != nil {
		return err
	}

	portMap := nat.PortMap{}

	utils.AppendIfExplicit(portMap, utils.PortSSH, cmd.Flags(), "ssh-port")
	utils.AppendIfExplicit(portMap, utils.PortVNC, cmd.Flags(), "vnc-port")
	utils.AppendIfExplicit(portMap, utils.PortAPI, cmd.Flags(), "k8s-port")
	utils.AppendIfExplicit(portMap, utils.PortOCP, cmd.Flags(), "ocp-port")
	utils.AppendIfExplicit(portMap, utils.PortRegistry, cmd.Flags(), "registry-port")

	qemuArgs, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	secondaryNics, err := cmd.Flags().GetUint("secondary-nics")
	if err != nil {
		return err
	}

	registryVol, err := cmd.Flags().GetString("registry-volume")
	if err != nil {
		return err
	}

	nfsData, err := cmd.Flags().GetString("nfs-data")
	if err != nil {
		return err
	}

	logDir, err := cmd.Flags().GetString("log-to-dir")
	if err != nil {
		return err
	}

	dockerProxy, err := cmd.Flags().GetString("docker-proxy")
	if err != nil {
		return err
	}

	cephEnabled, err := cmd.Flags().GetBool("enable-ceph")
	if err != nil {
		return err
	}

	cluster := args[0]

	background, err := cmd.Flags().GetBool("background")
	if err != nil {
		return err
	}

	containerRegistry, err := cmd.Flags().GetString("container-registry")
	if err != nil {
		return err
	}

	ovsdpdkEnabled, err := cmd.Flags().GetBool("enable-ovsdpdk")
	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	b := context.Background()
	ctx, cancel := context.WithCancel(b)

	containers, volumes, done := docker.NewCleanupHandler(cli, cmd.OutOrStderr())

	defer func() {
		done <- err
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		cancel()
		done <- fmt.Errorf("Interrupt received, clean up")
	}()

	if len(containerRegistry) > 0 {
		imageRef := path.Join(containerRegistry, cluster)
		fmt.Printf("Download the image %s\n", imageRef)
		err = docker.ImagePull(cli, ctx, imageRef, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
	}

	// Start dnsmasq
	dnsmasq, err := cli.ContainerCreate(ctx, &container.Config{
		Image: cluster,
		Env: []string{
			fmt.Sprintf("NUM_NODES=%d", nodes),
			fmt.Sprintf("NUM_SECONDARY_NICS=%d", secondaryNics),
		},
		Cmd: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
		ExposedPorts: nat.PortSet{
			utils.TCPPortOrDie(utils.PortSSH):      {},
			utils.TCPPortOrDie(utils.PortRegistry): {},
			utils.TCPPortOrDie(utils.PortOCP):      {},
			utils.TCPPortOrDie(utils.PortAPI):      {},
			utils.TCPPortOrDie(utils.PortVNC):      {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: randomPorts,
		PortBindings:    portMap,
		ExtraHosts: []string{
			"nfs:192.168.66.2",
			"registry:192.168.66.2",
			"ceph:192.168.66.2",
		},
	}, nil, prefix+"-dnsmasq")
	if err != nil {
		return err
	}
	containers <- dnsmasq.ID
	if err := cli.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Pull the registry image
	err = docker.ImagePull(cli, ctx, utils.DockerRegistryImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	// Create registry volume
	var registryMounts []mount.Mount
	if registryVol != "" {

		vol, err := cli.VolumeCreate(ctx, volume.VolumesCreateBody{
			Name: fmt.Sprintf("%s-%s", prefix, "registry"),
		})
		if err != nil {
			return err
		}
		registryMounts = []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: vol.Name,
				Target: "/var/lib/registry",
			},
		}
	}

	// Start registry
	registry, err := cli.ContainerCreate(ctx, &container.Config{
		Image: utils.DockerRegistryImage,
	}, &container.HostConfig{
		Mounts:      registryMounts,
		Privileged:  true, // fixme we just need proper selinux volume labeling
		NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
	}, nil, prefix+"-registry")
	if err != nil {
		return err
	}
	containers <- registry.ID
	if err := cli.ContainerStart(ctx, registry.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	if nfsData != "" {
		nfsData, err := filepath.Abs(nfsData)
		if err != nil {
			return err
		}
		// Pull the ganesha image
		err = docker.ImagePull(cli, ctx, utils.NFSGaneshaImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		// Start the ganesha image
		nfsServer, err := cli.ContainerCreate(ctx, &container.Config{
			Image: utils.NFSGaneshaImage,
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: nfsData,
					Target: "/data/nfs",
				},
			},
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
		}, nil, prefix+"-nfs-ganesha")
		if err != nil {
			return err
		}
		containers <- nfsServer.ID
		if err := cli.ContainerStart(ctx, nfsServer.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	}

	if cephEnabled {
		// Pull Ceph image
		err = docker.ImagePull(cli, ctx, utils.CephImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		cephStorage, err := cli.ContainerCreate(ctx, &container.Config{
			Image: utils.CephImage,
			Env: []string{
				"MON_IP=192.168.66.2",
				"CEPH_PUBLIC_NETWORK=0.0.0.0/0",
				"DEMO_DAEMONS=osd,mds",
				"CEPH_DEMO_UID=demo",
			},
			Cmd: strslice.StrSlice{
				"demo",
			},
		}, &container.HostConfig{
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
		}, nil, prefix+"-ceph")
		if err != nil {
			return err
		}
		containers <- cephStorage.ID
		if err := cli.ContainerStart(ctx, cephStorage.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	}

	if logDir != "" {
		logDir, err := filepath.Abs(logDir)
		if err != nil {
			return err
		}

		if _, err = os.Stat(logDir); os.IsNotExist(err) {
			os.Mkdir(logDir, 0755)
		}

		// Pull the fluent image
		err = docker.ImagePull(cli, ctx, utils.FluentdImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		// Start the fluent image
		fluentd, err := cli.ContainerCreate(ctx, &container.Config{
			Image: utils.FluentdImage,
			Cmd: strslice.StrSlice{
				"exec fluentd",
				"-i \"<system>\n log_level debug\n</system>\n<source>\n@type  forward\n@log_level error\nport  24224\n</source>\n<match **>\n@type file\npath /fluentd/log/collected\n</match>\"",
				"-p /fluentd/plugins $FLUENTD_OPT -v",
			},
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: logDir,
					Target: "/fluentd/log/collected",
				},
			},
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
		}, nil, prefix+"-fluentd")
		if err != nil {
			return err
		}
		containers <- fluentd.ID
		if err := cli.ContainerStart(ctx, fluentd.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	}

	// Add serial pty so we can do stuff like 'socat - /dev/pts0' to access
	// the VM console from the container without ssh
	qemuArgs += " -serial pty"

	wg := sync.WaitGroup{}
	wg.Add(int(nodes))
	// start one vm after each other
	macCounter := 0
	for x := 0; x < int(nodes); x++ {

		nodeQemuArgs := qemuArgs

		for i := 0; i < int(secondaryNics); i++ {
			netSuffix := fmt.Sprintf("%d-%d", x, i)
			macSuffix := fmt.Sprintf("%02x", macCounter)
			macCounter++
			nodeQemuArgs = fmt.Sprintf("%s -device virtio-net-pci,netdev=secondarynet%s,mac=52:55:00:d1:56:%s -netdev tap,id=secondarynet%s,ifname=stap%s,script=no,downscript=no", nodeQemuArgs, netSuffix, macSuffix, netSuffix, netSuffix)
		}

		if len(nodeQemuArgs) > 0 {
			nodeQemuArgs = "--qemu-args \"" + nodeQemuArgs + "\""
		}

		nodeName := nodeNameFromIndex(x + 1)
		nodeNum := fmt.Sprintf("%02d", x+1)
		if reverse {
			nodeName = nodeNameFromIndex((int(nodes) - x))
			nodeNum = fmt.Sprintf("%02d", (int(nodes) - x))
		}

		vol, err := cli.VolumeCreate(ctx, volume.VolumesCreateBody{
			Name: fmt.Sprintf("%s-%s", prefix, nodeName),
		})
		if err != nil {
			return err
		}
		volumes <- vol.Name

		node, err := cli.ContainerCreate(ctx, &container.Config{
			Image: cluster,
			Env: []string{
				fmt.Sprintf("NODE_NUM=%s", nodeNum),
			},
			Volumes: map[string]struct{}{
				"/var/run/disk/": {},
			},
			Cmd: []string{"/bin/bash", "-c", fmt.Sprintf("/vm.sh -n /var/run/disk/disk.qcow2 --memory %s --cpu %s %s", memory, strconv.Itoa(int(cpu)), nodeQemuArgs)},
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   "volume",
					Source: vol.Name,
					Target: "/var/run/disk",
				},
			},
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
		}, nil, prefix+"-"+nodeName)
		if err != nil {
			return err
		}
		containers <- node.ID
		if err := cli.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}

		// Wait for vm start
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "while [ ! -f /ssh_ready ] ; do sleep 1; done"}, os.Stdout)
		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("checking for ssh.sh script for node %s failed", nodeName)
		}

		if dockerProxy != "" {
			//if dockerProxy has value, genterate a shell script`/script/docker-proxy.sh` which can be applied to set proxy settings
			proxyConfig, err := getDockerProxyConfig(dockerProxy)
			if err != nil {
				return fmt.Errorf("parsing proxy settings for node %s failed", nodeName)
			}
			success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("cat <<EOF >/scripts/docker-proxy.sh %s", proxyConfig)}, os.Stdout)
			if err != nil {
				return fmt.Errorf("write failed for proxy provision script for node %s", nodeName)
			}
			if success {
				success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo /bin/bash < /scripts/docker-proxy.sh")}, os.Stdout)
			}
		}

		//check if we have a special provision script
		success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("test -f /scripts/%s.sh", nodeName)}, os.Stdout)
		if err != nil {
			return fmt.Errorf("checking for matching provision script for node %s failed", nodeName)
		}

		if success {
			success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo /bin/bash < /scripts/%s.sh", nodeName)}, os.Stdout)
		} else {
			if ovsdpdkEnabled {
				err = nodePackageUpdate(cli, nodeContainer(prefix, nodeName))
				if err != nil {
					return err
				}

				err = installOvs(cli, nodeContainer(prefix, nodeName))
				if err != nil {
					return err
				}
			}

			success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "ssh.sh sudo /bin/bash < /scripts/nodes.sh"}, os.Stdout)
		}

		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("provisioning node %s failed", nodeName)
		}

		go func(id string) {
			cli.ContainerWait(ctx, id)
			wg.Done()
		}(node.ID)
	}

	if cephEnabled {
		keyRing := new(bytes.Buffer)
		success, err := docker.Exec(cli, nodeContainer(prefix, "ceph"), []string{
			"/bin/bash",
			"-c",
			"ceph auth print-key client.admin | base64",
		}, keyRing)
		if err != nil {
			return err
		}
		nodeName := nodeNameFromIndex(1)
		key := bytes.TrimSpace(keyRing.Bytes())
		success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{
			"/bin/bash",
			"-c",
			fmt.Sprintf("ssh.sh sudo sed -i \"s/replace-me/%s/g\" /tmp/ceph/ceph-secret.yaml", key),
		}, os.Stdout)
		if err != nil {
			return err
		}
		success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{
			"/bin/bash",
			"-c",
			"ssh.sh sudo /bin/bash < /scripts/ceph-csi.sh",
		}, os.Stdout)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("provisioning Ceph CSI failed")
		}
	}

	// If logging is enabled, deploy the default fluent logging
	if logDir != "" {
		nodeName := nodeNameFromIndex(1)
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{
			"/bin/bash",
			"-c",
			"ssh.sh sudo /bin/bash < /scripts/logging.sh",
		}, os.Stdout)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("provisioning logging failed")
		}
	}

	// If background flag was specified, we don't want to clean up if we reach that state
	if !background {
		wg.Wait()
		done <- fmt.Errorf("Done. please clean up")
	}

	return nil
}

func nodeNameFromIndex(x int) string {
	return fmt.Sprintf("node%02d", x)
}

func nodeContainer(prefix string, node string) string {
	return prefix + "-" + node
}

func getDockerProxyConfig(proxy string) (string, error) {
	p := dockerSetting{Proxy: proxy}
	buf := new(bytes.Buffer)

	t, err := template.New("docker-proxy").Parse(proxySettings)
	if err != nil {
		return "", err
	}
	err = t.Execute(buf, p)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func runSSHCommand(cli *client.Client, nodeContainer, cmdString string) error {
	cmdString = "ssh.sh " + cmdString
	fmt.Println("Running cmd on node", nodeContainer, " - ", cmdString)
	success, err := docker.Exec(cli, nodeContainer, []string{"/bin/bash", "-c", cmdString}, os.Stdout)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("running command %s on node %s failed", cmdString, nodeContainer)
	}
	return nil
}

func nodePackageUpdate(cli *client.Client, nodeContainer string) error {
	var err error = nil
	err = runSSHCommand(cli, nodeContainer, "sudo /bin/bash < /scripts/kargs.sh")
	if err != nil {
		return err
	}

	err = runSSHCommand(cli, nodeContainer, "sudo shutdown -r 1")
	if err != nil {
		return err
	}

	fmt.Println("Rebooting node after update..")
	time.Sleep(2 * 60 * time.Second)

	// Wait for vm re-start
	success, err := docker.Exec(cli, nodeContainer, []string{"/bin/bash", "-c", "while [ ! -f /ssh_ready ] ; do sleep 1; done"}, os.Stdout)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("checking for ssh.sh script for node %s failed", nodeContainer)
	}

	return nil
}

func installOvs(cli *client.Client, nodeContainer string) error {
	// Install OvS
	success, err := docker.Exec(cli, nodeContainer, []string{"/bin/bash", "-c", "ssh.sh sudo /bin/bash < /scripts/ovsdpdk.sh"}, os.Stdout)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("installing ovs in node %s failed", nodeContainer)
	}
	return nil
}
