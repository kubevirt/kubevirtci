package libssh

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

//go:embed key.pem
var sshKey []byte

// Represents an interface to run a command on a node in the kubevirt cluster, the interface assumes only the bare command or script
// is going to be passed. any leading ways to configure the script like /bin/bash or anything is left to the caller to account for as an implementation detail
type Client interface {
	Command(cmd string) error
	CommandWithNoStdOut(cmd string) (string, error)
	CopyRemoteFile(remotePathToCopy string, out io.Writer) error
	SCP(destPath string, contents io.Reader) error
}

// Represents an interface to run a command on a node in the kubevirt cluster
// Implementation to the SSHClient interface based on native golang libraries
type SSHClientImpl struct {
	sshPort   uint16
	nodeIdx   int
	initMutex sync.Mutex
	config    *ssh.ClientConfig
	client    *ssh.Client
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
		config:    c,
		sshPort:   port,
		initMutex: sync.Mutex{},
		nodeIdx:   idx,
	}, nil
}

func (s *SSHClientImpl) Command(cmd string) error {
	return s.executeCommand(cmd, os.Stdout, os.Stderr)
}

func (s *SSHClientImpl) CommandWithNoStdOut(cmd string) (string, error) {
	var stdout, stderr bytes.Buffer

	err := s.executeCommand(cmd, &stdout, &stderr)
	if err != nil {
		return "", fmt.Errorf("%w, %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// Copies a file from a jump host after first establishing a connection with the forwarded port by dnsmasq
func (s *SSHClientImpl) SCP(fileName string, contents io.Reader) error {
	if s.client == nil {
		err := s.initClient()
		if err != nil {
			return err
		}
	}

	scpClient, err := scp.NewClientBySSH(s.client)
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

// Copies a file on a jump host after first establishing a connection with the forwarded port by dnsmasq
func (s *SSHClientImpl) CopyRemoteFile(remotePathToCopy string, target io.Writer) error {
	if s.client == nil {
		err := s.initClient()
		if err != nil {
			return err
		}
	}

	session, err := s.client.NewSession()
	if err != nil {
		return err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stdout for session: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stderr for session: %v", err)
	}
	go io.Copy(os.Stderr, stderr)

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

	err = session.Run("sudo -i /usr/bin/scp -qf " + remotePathToCopy)

	copyError := <-errChan

	if err == nil && copyError != nil {
		return copyError
	}

	return err
}

func (s *SSHClientImpl) executeCommand(cmd string, outWriter, errWriter io.Writer) error {
	if s.client == nil {
		err := s.initClient()
		if err != nil {
			return err
		}
	}
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = outWriter
	session.Stderr = errWriter

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
		return fmt.Errorf("failed to execute command: %s", cmd)
	}
	return nil
}

func (s *SSHClientImpl) initClient() error {
	s.initMutex.Lock()
	defer s.initMutex.Unlock()
	client, err := ssh.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(s.sshPort)), s.config)
	if err != nil {
		return fmt.Errorf("Failed to connect to SSH server: %v", err)
	}

	conn, err := client.Dial("tcp", fmt.Sprintf("192.168.66.10%d:22", s.nodeIdx))
	if err != nil {
		return fmt.Errorf("Error establishing connection to the next hop host: %s", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, fmt.Sprintf("192.168.66.10%d:22", s.nodeIdx), s.config)
	if err != nil {
		return fmt.Errorf("Error creating forwarded ssh connection: %s", err)
	}

	jumpHost := ssh.NewClient(ncc, chans, reqs)
	s.client = jumpHost
	return nil
}
