package podman

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
)

type Podman struct{}

func NewPodman() *Podman {
	return &Podman{}
}

type PodmanSSHClient struct {
	containerName string
}

func NewPodmanSSHClient(containerName string) *PodmanSSHClient {
	return &PodmanSSHClient{
		containerName: containerName,
	}
}

func IsAvailable() bool {
	cmd := exec.Command("podman", "-v")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.HasPrefix(string(out), "podman version")
}

func (p *PodmanSSHClient) Command(cmd string) error {
	logrus.Infof("[node %s]: %s\n", p.containerName, cmd)
	command := exec.Command("podman", "exec", p.containerName, "/bin/sh", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

func (p *PodmanSSHClient) CommandWithNoStdOut(cmd string) (string, error) {
	command := exec.Command("podman", "exec", p.containerName, "/bin/sh", "-c", cmd)
	out, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (p *PodmanSSHClient) CopyRemoteFile(remotePath string, out io.Writer) error {
	defer os.Remove("tempfile")
	cmd := exec.Command("podman", "cp", fmt.Sprintf("%s:%s", p.containerName, remotePath), "tempfile")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy file from container: %w, output: %s", err, output)
	}

	tmpfile, err := os.ReadFile("tempfile")
	if err != nil {
		return err
	}

	_, err = out.Write(tmpfile)
	if err != nil {
		return err
	}

	return nil
}

func (p *PodmanSSHClient) SCP(destPath string, contents io.Reader) error {
	tempFile, err := os.CreateTemp("", "podman_cp_temp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	fileContents, err := io.ReadAll(contents)
	if err != nil {
		return fmt.Errorf("failed to read file contents: %w", err)
	}

	_, err = tempFile.Write(fileContents)
	if err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	err = tempFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	cmd := exec.Command("podman", "cp", tempFile.Name(), p.containerName+":"+destPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("podman cp command failed: %w. Output: %s", err, string(output))
	}

	return nil
}

func (p *Podman) ImagePull(image string) error {
	cmd := exec.Command("podman", "pull", image)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (p *Podman) Create(image string, createOpts *cri.CreateOpts) (string, error) {
	ports := ""
	for containerPort, hostPort := range createOpts.Ports {
		ports += "-p " + containerPort + ":" + hostPort
	}

	args := []string{
		"--name=" + createOpts.Name,
		"--privileged=" + strconv.FormatBool(createOpts.Privileged),
		"--rm=" + strconv.FormatBool(createOpts.Remove),
		"--restart=" + createOpts.RestartPolicy,
		"--network=" + createOpts.Network,
	}

	for containerPort, hostPort := range createOpts.Ports {
		args = append(args, "-p", containerPort+":"+hostPort)
	}

	if len(createOpts.Capabilities) > 0 {
		args = append(args, "--cap-add="+strings.Join(createOpts.Capabilities, ","))
	}

	fullArgs := append([]string{"create"}, args...)
	fullArgs = append(fullArgs, image)
	fullArgs = append(fullArgs, createOpts.Command...)

	cmd := exec.Command("podman",
		fullArgs...,
	)
	fmt.Println(cmd.String())

	containerID, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	logrus.Info("created registry container with id: ", string(containerID))
	return strings.TrimSuffix(string(containerID), "\n"), nil
}

func (p *Podman) Start(containerID string) error {
	cmd := exec.Command("podman",
		"start",
		containerID)

	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func (p *Podman) Inspect(containerID, format string) ([]byte, error) {
	cmd := exec.Command("podman", "inspect", containerID, "--format", format)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (p *Podman) Remove(containerID string) error {
	cmd := exec.Command("podman", "rm", "-f", containerID)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}