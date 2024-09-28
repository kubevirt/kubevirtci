package cmd

import (
	"github.com/spf13/cobra"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/kindbase"
)

func NewRemoveKindCommand() *cobra.Command {
	rm := &cobra.Command{
		Use:   "rm-kind",
		Short: "rm deletes all traces of a kind cluster",
		RunE:  rmKind,
		Args:  cobra.ExactArgs(1),
	}
	return rm
}

func rmKind(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	kindProvider, err := kind.NewKindBaseProvider(&kind.KindConfig{Version: prefix})
	if err != nil {
		return err
	}
	if err = kindProvider.Delete(); err != nil {
		return err
	}

	return nil
}
