// +build !windows

package varlink

import (
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/varlink/go/varlink/internal/ctxio"
)

type PipeCon struct {
	net.Conn
	cmd    *exec.Cmd
	reader *io.ReadCloser
	writer *io.WriteCloser
}

func (p PipeCon) Close() error {
	err1 := (*p.reader).Close()
	err2 := (*p.writer).Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	p.cmd.Wait()

	return nil
}

// NewBridgeWithStderr returns a new connection with the given bridge.
func NewBridgeWithStderr(bridge string, stderr io.Writer) (*Connection, error) {
	c := Connection{}
	cmd := exec.Command("sh", "-c", bridge)
	cmd.Stderr = stderr
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	c.conn = ctxio.NewConn(PipeCon{nil, cmd, &r, &w})
	c.address = ""

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// NewBridge returns a new connection with the given bridge.
func NewBridge(bridge string) (*Connection, error) {
	return NewBridgeWithStderr(bridge, os.Stderr)
}
