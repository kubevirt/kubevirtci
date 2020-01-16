package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fromanirh/pack8s/cmd/cmdutil"
)

type showOptions struct {
	containerIdsOnly bool
}

var showOpts showOptions

func NewShowCommand() *cobra.Command {
	show := &cobra.Command{
		Use:   "show",
		Short: "show lists containers belonging to the cluster",
		RunE:  showContainers,
		Args:  cobra.NoArgs,
	}

	show.Flags().BoolVarP(&showOpts.containerIdsOnly, "ids", "i", false, "show only container ids")

	return show
}

/*
 * The show command is unique among the other pack8s commands because it is the only one it wants
 * to display informations on stdout.
 * Hence, we make explicit use of fmt.*[Pp]rintf and not just logging.
 * The output here is meant for the immediate and actual consumption of the user: that's the result
 * the user asked and wanted to have.
 * Logging OTOH is rarely explciitely asked by the user, and always as by product because something
 * else didn't go according to the plan.
 */
func showContainers(cmd *cobra.Command, args []string) error {
	cOpts, err := cmdutil.GetCommonOpts(cmd)
	if err != nil {
		return err
	}

	hnd, _, err := cOpts.GetHandle()
	if err != nil {
		return err
	}

	containers, err := hnd.GetPrefixedContainers(cOpts.Prefix)
	if err != nil {
		return err
	}

	if showOpts.containerIdsOnly {
		for _, cont := range containers {
			fmt.Printf("%s\n", cont.Id)
		}
		return nil
	} else {
		if len(containers) >= 1 {
			fmt.Printf("# Container:\n")
			for _, cont := range containers {
				fmt.Printf("%-32s\t%s\n", cont.Names, cont.Id)
			}
		} else {
			fmt.Printf("no containers found for cluster %s\n", cOpts.Prefix)
		}
	}

	volumes, err := hnd.GetPrefixedVolumes(cOpts.Prefix)
	if err != nil {
		return err
	}

	if len(volumes) >= 1 {
		fmt.Printf("# Volumes:\n")
		for _, vol := range volumes {
			fmt.Printf("%-32s\t@%s\n", vol.Name, vol.MountPoint)
		}
	} else {
		fmt.Printf("no volumes found for cluster %s\n", cOpts.Prefix)
	}

	return nil
}
