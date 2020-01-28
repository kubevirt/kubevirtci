package podman

import (
	"context"

	"github.com/varlink/go/varlink"

	"github.com/fromanirh/pack8s/iopodman"
)

// ExecContainer executes a command in the given container.
type ExecContainer_methods struct{}

func ExecContainer() ExecContainer_methods { return ExecContainer_methods{} }

func (m ExecContainer_methods) Call(ctx context.Context, c *varlink.Connection, opts_in_ iopodman.ExecOpts) (rwc_ varlink.ReadWriterContext, err_ error) {
	receive, err_ := m.Upgrade(ctx, c, opts_in_)
	if err_ != nil {
		return
	}
	rwc_, err_ = receive(ctx)
	return
}

func (m ExecContainer_methods) Upgrade(ctx context.Context, c *varlink.Connection, opts_in_ iopodman.ExecOpts) (func(ctx context.Context) (varlink.ReadWriterContext, error), error) {
	var in struct {
		Opts iopodman.ExecOpts `json:"opts"`
	}
	in.Opts = opts_in_
	receive, err := c.Upgrade(ctx, "io.podman.ExecContainer", in)
	if err != nil {
		return nil, err
	}
	return func(context.Context) (rwc varlink.ReadWriterContext, err error) {
		_, rwc, err = receive(ctx, nil)
		if err != nil {
			err = iopodman.Dispatch_Error(err)
			return
		}
		return
	}, nil
}
