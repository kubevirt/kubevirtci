package cmd

import (
	"github.com/spf13/cobra"

	"github.com/fromanirh/pack8s/cmd/cmdutil"
)

// NewPruneVolumesCommand returns command to prune unused volumes
func NewPruneVolumesCommand() *cobra.Command {
	prune := &cobra.Command{
		Use:   "prune",
		Short: "prune removes unused volumes on the host",
		RunE:  pruneVolumes,
		Args:  cobra.NoArgs,
	}
	return prune
}

func pruneVolumes(cmd *cobra.Command, _ []string) error {
	cOpts, err := cmdutil.GetCommonOpts(cmd)
	if err != nil {
		return err
	}

	hnd, _, err := cOpts.GetHandle()
	if err != nil {
		return err
	}

	return hnd.PruneVolumes()
}
