package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"os/signal"
)

func GetPrefixedContainers(cli *client.Client, prefix string) ([]types.Container, error) {
	args, err := filters.ParseFlag("name="+prefix, filters.NewArgs())
	if err != nil {
		return nil, err
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: args,
		All:     true,
	})
	return containers, err
}

func GetPrefixedVolumes(cli *client.Client, prefix string) ([]*types.Volume, error) {
	args, err := filters.ParseFlag("name="+prefix, filters.NewArgs())
	if err != nil {
		return nil, err
	}
	volumes, err := cli.VolumeList(context.Background(), args)
	if err != nil {
		return nil, err
	}
	return volumes.Volumes, nil
}

func GetDDNSMasqContainer(cli *client.Client, prefix string) (*types.Container, error) {
	containers, err := GetPrefixedContainers(cli, prefix+"-"+"dnsmasq")
	if err != nil {
		return nil, err
	}

	if len(containers) == 1 {
		return &containers[0], nil
	}

	return nil, fmt.Errorf("Could not identify dnsmasq container %s", prefix+"-dnsmasq")
}

func Exec(cli *client.Client, container string, args []string, out io.Writer) (bool, error) {
	ctx := context.Background()
	id, err := cli.ContainerExecCreate(ctx, container, types.ExecConfig{
		Privileged:   true,
		Tty:          false,
		Detach:       false,
		Cmd:          args,
		AttachStdout: true,
		AttachStderr: true,
	})

	if err != nil {
		return false, err
	}

	attached, err := cli.ContainerExecAttach(ctx, id.ID, types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Tty:          false,
	})
	if err != nil {
		return false, err
	}
	defer attached.Close()

	io.Copy(out, attached.Reader)

	resp, err := cli.ContainerExecInspect(ctx, id.ID)
	if err != nil {
		return false, err
	}
	return resp.ExitCode == 0, nil
}

func Terminal(cli *client.Client, container string, args []string, file *os.File) (int, error) {

	ctx := context.Background()
	id, err := cli.ContainerExecCreate(ctx, container, types.ExecConfig{
		Privileged:   true,
		Tty:          terminal.IsTerminal(int(file.Fd())),
		Detach:       false,
		Cmd:          args,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
	})

	if err != nil {
		return -1, err
	}

	attached, err := cli.ContainerExecAttach(ctx, id.ID, types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  true,
		Tty:          terminal.IsTerminal(int(file.Fd())),
	})
	if err != nil {
		return -1, err
	}
	defer attached.Close()

	if terminal.IsTerminal(int(file.Fd())) {
		state, err := terminal.MakeRaw(int(file.Fd()))
		if err != nil {
			return -1, err
		}

		errChan := make(chan error)

		go func() {
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)
			<-interrupt
			close(errChan)
		}()

		go func() {
			_, err := io.Copy(file, attached.Conn)
			errChan <- err
		}()

		go func() {
			_, err := io.Copy(attached.Conn, file)
			errChan <- err
		}()

		defer func() {
			terminal.Restore(int(file.Fd()), state)
		}()

		err = <-errChan

		if err != nil {
			return -1, err
		}
	}

	resp, err := cli.ContainerExecInspect(ctx, id.ID)
	if err != nil {
		return -1, err
	}
	return resp.ExitCode, nil
}
