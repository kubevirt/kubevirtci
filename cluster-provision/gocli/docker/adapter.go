package docker

import (
	"fmt"
	"os"

	"bytes"
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// DockerAdapter is a wrapper around client.Client to conform it to the SSH interface
type DockerAdapter struct {
	nodeName     string
	dockerClient *client.Client
}

func NewAdapter(cli *client.Client, nodeName string) *DockerAdapter {
	return &DockerAdapter{
		nodeName:     nodeName,
		dockerClient: cli,
	}
}

func (d *DockerAdapter) Command(cmd string) error {
	success, err := Exec(d.dockerClient, d.nodeName, []string{"/bin/sh", "-c", cmd}, os.Stdout)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("Error executing %s on node %s", cmd, d.nodeName)
	}
	return nil
}

func (d *DockerAdapter) CommandWithNoStdOut(cmd string) (string, error) {
	var buf *bytes.Buffer
	success, err := Exec(d.dockerClient, d.nodeName, []string{"/bin/sh", "-c", cmd}, buf)
	if err != nil {
		return "", err
	}

	if !success {
		return "", fmt.Errorf("Error executing %s on node %s", cmd, d.nodeName)
	}
	return buf.String(), nil
}

func (d *DockerAdapter) SCP(destPath string, contents io.Reader) error {
	return d.dockerClient.CopyToContainer(context.Background(), d.nodeName, destPath, contents, types.CopyToContainerOptions{})
}

func (d *DockerAdapter) CopyRemoteFile(remotePath string, out io.Writer) error {
	defer os.Remove("tempfile")
	if _, _, err := d.dockerClient.CopyFromContainer(context.Background(), d.nodeName, "tempfile"); err != nil {
		return err
	}

	tempfile, err := os.ReadFile("tempfile")
	if err != nil {
		return err
	}

	_, err = out.Write(tempfile)
	if err != nil {
		return err
	}
	return nil
}
