package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	scp1 "github.com/bramvdbogaerde/go-scp"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	ssh1 "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

// NewSSHCommand returns command to SSH to the cluster node
func NewSSHCommand() *cobra.Command {

	ssh := &cobra.Command{
		Use:   "ssh",
		Short: "ssh into a node",
		RunE:  ssh,
		Args:  cobra.MinimumNArgs(1),
	}
	return ssh
}

func ssh(cmd *cobra.Command, args []string) error {
	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	node := args[0]

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	// TODO we can do the ssh session with the native golang client
	container := prefix + "-" + node
	ssh_command := append([]string{"ssh.sh"}, args[1:]...)
	file := os.Stdout
	if terminal.IsTerminal(int(file.Fd())) {
		exitCode, err := docker.Terminal(cli, container, ssh_command, file)
		if err != nil {
			return err
		}
		os.Exit(exitCode)
	} else {
		execExitCodeIsZero, err := docker.Exec(cli, container, ssh_command, file)
		if err != nil {
			return err
		}
		exitCode := 0
		if !execExitCodeIsZero {
			exitCode = 1
		}
		os.Exit(exitCode)
	}
	return nil
}

func jumpSSH(sshPort uint16, nodeIdx int, cmd string, stdOut bool) (string, error) {
	signer, err := ssh1.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return "", err
	}

	config := &ssh1.ClientConfig{
		User: "vagrant",
		Auth: []ssh1.AuthMethod{
			ssh1.PublicKeys(signer),
		},
		HostKeyCallback: ssh1.InsecureIgnoreHostKey(),
	}

	client, err := ssh1.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(sshPort)), config)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	conn, err := client.Dial("tcp", fmt.Sprintf("192.168.66.10%d:22", nodeIdx))
	if err != nil {
		return "", fmt.Errorf("Error establishing connection to the next hop host: %s", err)
	}

	ncc, chans, reqs, err := ssh1.NewClientConn(conn, fmt.Sprintf("192.168.66.10%d:22", nodeIdx), config)
	if err != nil {
		return "", fmt.Errorf("Error creating forwarded ssh connection: %s", err)
	}
	jumpHost := ssh1.NewClient(ncc, chans, reqs)
	session, err := jumpHost.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	session.Stderr = &stderr
	session.Stdout = os.Stdout
	if !stdOut {
		session.Stdout = &stdout
	}

	err = session.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("Failed to execute command: %v, %v", err, stderr.String())
	}
	return stdout.String(), nil
}

func jumpSCP(sshPort uint16, destNodeIdx int, fileName string) error {
	signer, err := ssh1.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return err
	}

	config := &ssh1.ClientConfig{
		User: "vagrant",
		Auth: []ssh1.AuthMethod{
			ssh1.PublicKeys(signer),
		},
		HostKeyCallback: ssh1.InsecureIgnoreHostKey(),
	}

	client, err := ssh1.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(sshPort)), config)
	if err != nil {
		return fmt.Errorf("Failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	conn, err := client.Dial("tcp", fmt.Sprintf("192.168.66.10%d:22", destNodeIdx))
	if err != nil {
		return fmt.Errorf("Error establishing connection to the next hop host: %s", err)
	}

	ncc, chans, reqs, err := ssh1.NewClientConn(conn, fmt.Sprintf("192.168.66.10%d:22", destNodeIdx), config)
	if err != nil {
		return fmt.Errorf("Error creating forwarded ssh connection: %s", err)
	}

	jumpHost := ssh1.NewClient(ncc, chans, reqs)
	defer jumpHost.Close()

	scpClient, err := scp1.NewClientBySSH(jumpHost)
	if err != nil {
		return err
	}

	err = scpClient.Connect()
	if err != nil {
		return err
	}

	file, err := f.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	filename := strings.Split(fileName, "/")
	err = scpClient.CopyFile(context.Background(), file, "/home/vagrant/"+filename[len(filename)-1], "0775")
	if err != nil {
		return err
	}

	return nil
}

func copyRemoteFile(sshPort uint16, remotePath, localPath string) error {
	signer, err := ssh1.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return err
	}

	config := &ssh1.ClientConfig{
		User: "vagrant",
		Auth: []ssh1.AuthMethod{
			ssh1.PublicKeys(signer),
		},
		HostKeyCallback: ssh1.InsecureIgnoreHostKey(),
	}

	connection, err := ssh1.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", sshPort), config)
	if err != nil {
		return err
	}

	session, err := connection.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stdout for session: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stderr for session: %v", err)
	}
	go io.Copy(os.Stderr, stderr)

	var target *os.File
	if localPath == "-" {
		target = os.Stdout
	} else {
		target, err = os.Create(localPath)
		if err != nil {
			return err
		}
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		b := make([]byte, 1)
		var buf bytes.Buffer
		for {
			n, err := stdout.Read(b)
			if err != nil {
				errChan <- fmt.Errorf("error: %v", err)
				return
			}
			if n == 0 {
				continue
			}

			if b[0] == '\n' {
				break
			}
			buf.WriteByte(b[0])
		}

		metadata := strings.Split(buf.String(), " ")
		if len(metadata) < 3 || !strings.HasPrefix(buf.String(), "C") {
			errChan <- fmt.Errorf("%v", buf.String())
			return
		}
		l, err := strconv.Atoi(metadata[1])
		if err != nil {
			errChan <- fmt.Errorf("invalid metadata: %v", buf.String())
			return
		}
		_, err = io.CopyN(target, stdout, int64(l))
		errChan <- err
	}()
	wrPipe, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to open pipe: %v", err)
	}

	go func(wrPipe io.WriteCloser) {
		defer wrPipe.Close()
		fmt.Fprintf(wrPipe, "\x00")
		fmt.Fprintf(wrPipe, "\x00")
		fmt.Fprintf(wrPipe, "\x00")
		fmt.Fprintf(wrPipe, "\x00")
	}(wrPipe)

	err = session.Run("sudo -i /usr/bin/scp -qf " + remotePath)

	copyError := <-errChan

	if err == nil && copyError != nil {
		return copyError
	}

	return err
}
