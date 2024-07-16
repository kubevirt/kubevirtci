package libssh

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

//go:embed key.pem
var sshKey []byte

// Represents an interface to run a command on a node in the kubevirt cluster, the interface assumes only the bare command or script
// is going to be passed. any leading ways to configure the script like /bin/bash or anything is left to the caller to account for as an implementation detail
type Client interface {
	Command(cmd string, stdOut bool) (string, error)
	CopyRemoteFile(remotePath, localPath string) error
}

// Represents an interface to run a command on a node in the kubevirt cluster

// Implementation to the SSHClient interface based on native golang libraries
type SSHClientImpl struct {
	sshPort uint16
	nodeIdx int
	config  *ssh.ClientConfig
}

func NewSSHClient(port uint16, idx int, root bool) (*SSHClientImpl, error) {
	signer, err := ssh.ParsePrivateKey(sshKey)
	if err != nil {
		return nil, err
	}
	u := "vagrant"
	if root {
		u = "root"
	}

	c := &ssh.ClientConfig{
		User: u,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return &SSHClientImpl{
		config:  c,
		sshPort: port,
		nodeIdx: idx,
	}, nil
}

// SSH performs two ssh connections, one to the forwarded port by dnsmasq to the local which is the ssh port of the control plane node
// then a hop to the designated host where the command is desired to be ran

func (s *SSHClientImpl) Command(cmd string, stdOut bool) (string, error) {
	client, err := ssh.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(s.sshPort)), s.config)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	conn, err := client.Dial("tcp", fmt.Sprintf("192.168.66.10%d:22", s.nodeIdx))
	if err != nil {
		return "", fmt.Errorf("Error establishing connection to the next hop host: %s", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("192.168.66.10%d:22", s.nodeIdx), s.config)
	if err != nil {
		return "", fmt.Errorf("Error creating forwarded ssh connection: %s", err)
	}
	jumpHost := ssh.NewClient(ncc, chans, reqs)
	session, err := jumpHost.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var stdout, stderr bytes.Buffer

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	if !stdOut {
		session.Stdout = &stdout
		session.Stderr = &stderr
	}

	if len(cmd) > 0 {
		firstCmdChar := cmd[0]
		// indicates the command is a script or a script with params
		if string(firstCmdChar) == "/" || string(firstCmdChar) == "-" {
			cmd = "sudo /bin/bash " + cmd
		}
	}
	logrus.Infof("[node %d]: %s", s.nodeIdx, cmd)

	err = session.Run(cmd)
	if err != nil {
		if !stdOut {
			err = fmt.Errorf(stderr.String())
		}
		return "", fmt.Errorf("Failed to execute command: %v, %v", cmd, err)
	}
	return stdout.String(), nil
}

func (s *SSHClientImpl) CopyRemoteFile(remotePath, localPath string) error {
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

	connection, err := ssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", s.sshPort), config)
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
