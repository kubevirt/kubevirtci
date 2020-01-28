package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	ssh1 "golang.org/x/crypto/ssh"

	"github.com/spf13/cobra"

	"github.com/fromanirh/pack8s/internal/pkg/ports"

	"github.com/fromanirh/pack8s/cmd/cmdutil"
)

const sshKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzI
w+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoP
kcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2
hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NO
Td0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcW
yLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQIBIwKCAQEA4iqWPJXtzZA68mKd
ELs4jJsdyky+ewdZeNds5tjcnHU5zUYE25K+ffJED9qUWICcLZDc81TGWjHyAqD1
Bw7XpgUwFgeUJwUlzQurAv+/ySnxiwuaGJfhFM1CaQHzfXphgVml+fZUvnJUTvzf
TK2Lg6EdbUE9TarUlBf/xPfuEhMSlIE5keb/Zz3/LUlRg8yDqz5w+QWVJ4utnKnK
iqwZN0mwpwU7YSyJhlT4YV1F3n4YjLswM5wJs2oqm0jssQu/BT0tyEXNDYBLEF4A
sClaWuSJ2kjq7KhrrYXzagqhnSei9ODYFShJu8UWVec3Ihb5ZXlzO6vdNQ1J9Xsf
4m+2ywKBgQD6qFxx/Rv9CNN96l/4rb14HKirC2o/orApiHmHDsURs5rUKDx0f9iP
cXN7S1uePXuJRK/5hsubaOCx3Owd2u9gD6Oq0CsMkE4CUSiJcYrMANtx54cGH7Rk
EjFZxK8xAv1ldELEyxrFqkbE4BKd8QOt414qjvTGyAK+OLD3M2QdCQKBgQDtx8pN
CAxR7yhHbIWT1AH66+XWN8bXq7l3RO/ukeaci98JfkbkxURZhtxV/HHuvUhnPLdX
3TwygPBYZFNo4pzVEhzWoTtnEtrFueKxyc3+LjZpuo+mBlQ6ORtfgkr9gBVphXZG
YEzkCD3lVdl8L4cw9BVpKrJCs1c5taGjDgdInQKBgHm/fVvv96bJxc9x1tffXAcj
3OVdUN0UgXNCSaf/3A/phbeBQe9xS+3mpc4r6qvx+iy69mNBeNZ0xOitIjpjBo2+
dBEjSBwLk5q5tJqHmy/jKMJL4n9ROlx93XS+njxgibTvU6Fp9w+NOFD/HvxB3Tcz
6+jJF85D5BNAG3DBMKBjAoGBAOAxZvgsKN+JuENXsST7F89Tck2iTcQIT8g5rwWC
P9Vt74yboe2kDT531w8+egz7nAmRBKNM751U/95P9t88EDacDI/Z2OwnuFQHCPDF
llYOUI+SpLJ6/vURRbHSnnn8a/XG+nzedGH5JGqEJNQsz+xT2axM0/W/CRknmGaJ
kda/AoGANWrLCz708y7VYgAtW2Uf1DPOIYMdvo6fxIB5i9ZfISgcJ/bbCUkFrhoH
+vq/5CIWxCPp0f85R4qxxQ5ihxJ0YDQT9Jpx4TMss4PSavPaBH3RXow5Ohe+bYoQ
NE5OgEXk2wVfZczCZpigBKbKZHNYcelXtTt/nP3rsCuGcM4h53s=
-----END RSA PRIVATE KEY-----`

type scpOptions struct {
	containerName string
	sshUser       string
}

var scpOpts scpOptions

// NewSCPCommand returns command to copy files via SSH from the cluster node to localhost
func NewSCPCommand() *cobra.Command {

	ssh := &cobra.Command{
		Use:   "scp SRC DST",
		Short: "scp copies files from master node to the local host",
		RunE:  scp,
		Args:  cobra.ExactArgs(2),
	}

	ssh.Flags().StringVar(&scpOpts.containerName, "container-name", "dnsmasq", "the container name to SSH copy from")
	ssh.Flags().StringVar(&scpOpts.sshUser, "ssh-user", "vagrant", "the user that used to connect via SSH to the node")

	return ssh
}

func scp(cmd *cobra.Command, args []string) error {
	cOpts, err := cmdutil.GetCommonOpts(cmd)
	if err != nil {
		return err
	}

	src := args[0]
	dst := args[1]

	hnd, log, err := cOpts.GetHandle()
	if err != nil {
		return err
	}

	cont, err := hnd.FindPrefixedContainer(cOpts.Prefix + "-" + scpOpts.containerName)
	if err != nil {
		return err
	}

	sshPort, err := ports.GetPublicPort(ports.PortSSH, cont.Ports)
	if err != nil {
		return err
	}

	log.Noticef("scp [%s:%d/%v] => [%v]", cont.Names, sshPort, src, dst)

	signer, err := ssh1.ParsePrivateKey([]byte(sshKey))
	if err != nil {
		return err
	}

	config := &ssh1.ClientConfig{
		User: scpOpts.sshUser,
		Auth: []ssh1.AuthMethod{
			ssh1.PublicKeys(signer),
		},
		HostKeyCallback: ssh1.InsecureIgnoreHostKey(),
	}

	sshAddr := fmt.Sprintf("127.0.0.1:%v", sshPort)
	connection, err := ssh1.Dial("tcp", sshAddr, config)
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
	if dst == "-" {
		target = os.Stdout
	} else {
		target, err = os.Create(dst)
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

	go func() {
		wrPipe, _ := session.StdinPipe()
		defer wrPipe.Close()

		fmt.Fprintf(wrPipe, "\x00")
		fmt.Fprintf(wrPipe, "\x00")
		fmt.Fprintf(wrPipe, "\x00")
		fmt.Fprintf(wrPipe, "\x00")
	}()

	err = session.Run("sudo -i /usr/bin/scp -qf " + src)

	copyError := <-errChan

	if err == nil && copyError != nil {
		log.Noticef("scp: copy error: %v", copyError)
		return copyError
	}

	if copyError != nil {
		fmt.Fprintln(cmd.OutOrStderr(), copyError)
		log.Noticef("scp: copy failed: %v", copyError)
	}

	log.Noticef("scp: copy done: (err=%v)", err)
	return err
}
