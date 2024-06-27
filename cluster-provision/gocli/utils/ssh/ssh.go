package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SSHClient interface {
	JumpSSH(uint16, int, string, bool, bool) (string, error)
	JumpSCP(uint16, int, string, fs.File) error
	CopyRemoteFile(uint16, string, string) error
}

type SSHClientImpl struct{}

func (s *SSHClientImpl) JumpSSH(sshPort uint16, nodeIdx int, cmd string, root, stdOut bool) (string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return "", err
	}
	u := "vagrant"
	if root {
		u = "root"
	}

	config := &ssh.ClientConfig{
		User: u,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(sshPort)), config)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	conn, err := client.Dial("tcp", fmt.Sprintf("192.168.66.10%d:22", nodeIdx))
	if err != nil {
		return "", fmt.Errorf("Error establishing connection to the next hop host: %s", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("192.168.66.10%d:22", nodeIdx), config)
	if err != nil {
		return "", fmt.Errorf("Error creating forwarded ssh connection: %s", err)
	}
	jumpHost := ssh.NewClient(ncc, chans, reqs)
	session, err := jumpHost.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	if !stdOut {
		session.Stdout = &stdout
		session.Stderr = &stderr
	}
	logrus.Infof("[node %d]: %s\n", nodeIdx, cmd)

	err = session.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("Failed to execute command: %v, %v", err, stderr.String())
	}
	return stdout.String(), nil
}

func (s *SSHClientImpl) JumpSCP(sshPort uint16, destNodeIdx int, fileName string, contents fs.File) error {
	signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: "vagrant",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(sshPort)), config)
	if err != nil {
		return fmt.Errorf("Failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	conn, err := client.Dial("tcp", fmt.Sprintf("192.168.66.10%d:22", destNodeIdx))
	if err != nil {
		return fmt.Errorf("Error establishing connection to the next hop host: %s", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("192.168.66.10%d:22", destNodeIdx), config)
	if err != nil {
		return fmt.Errorf("Error creating forwarded ssh connection: %s", err)
	}

	jumpHost := ssh.NewClient(ncc, chans, reqs)
	defer jumpHost.Close()

	scpClient, err := scp.NewClientBySSH(jumpHost)
	if err != nil {
		return err
	}

	err = scpClient.Connect()
	if err != nil {
		return err
	}

	err = scpClient.CopyFile(context.Background(), contents, fileName, "0775")
	if err != nil {
		return err
	}

	return nil
}

func (s *SSHClientImpl) CopyRemoteFile(sshPort uint16, remotePath, localPath string) error {
	signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: "vagrant",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	connection, err := ssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", sshPort), config)
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
