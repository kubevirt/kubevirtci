package cmd

import (
	"fmt"
	"os"
	"time"

	logger "github.com/apsdehal/go-logger"
	"github.com/spf13/cobra"
	spin "github.com/tj/go-spin"

	"github.com/fromanirh/pack8s/cmd/cmdutil"
)

type pullOptions struct {
	auxImages bool
}

func (po pullOptions) WantsNFS() bool {
	return po.auxImages
}

func (po pullOptions) WantsCeph() bool {
	return po.auxImages
}

func (po pullOptions) WantsFluentd() bool {
	return po.auxImages
}

var pullOpts pullOptions

func NewPullCommand() *cobra.Command {
	show := &cobra.Command{
		Use:   "pull",
		Short: "pull downloads an image from a registry",
		RunE:  pullImage,
		Args:  cobra.ExactArgs(1),
	}
	show.Flags().BoolVarP(&pullOpts.auxImages, "aux-images", "a", false, "pull the cluster auxiliary images")
	return show
}

type termProgressReporter struct {
	Log      *logger.Logger
	Spin     *spin.Spinner
	Interval time.Duration
}

func (tpp termProgressReporter) GetInterval() time.Duration {
	return tpp.Interval
}

func (tpp termProgressReporter) Report(ref string, elapsed, completed uint64, err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n")
		tpp.Log.Warningf("download failed for %s: %v\n", ref, err)
	} else if completed != 0 && elapsed == completed {
		fmt.Fprintf(os.Stderr, "\n")
		tpp.Log.Noticef("downloaded completed for %s", ref)
	} else {
		fmt.Fprintf(os.Stderr, "\rdownloading %s... %s ", ref, tpp.Spin.Next())
	}
	return err

}

func pullImage(cmd *cobra.Command, args []string) error {
	cOpts, err := cmdutil.GetCommonOpts(cmd)
	if err != nil {
		return err
	}

	hnd, log, err := cOpts.GetHandle()
	if err != nil {
		return err
	}

	if cOpts.IsTTY {
		s := spin.New()
		s.Set(spin.Spin1)
		tpp := termProgressReporter{
			Log:      log,
			Spin:     s,
			Interval: 1,
		}
		hnd.PullReporter = tpp
	}

	cluster := args[0]
	if pullOpts.auxImages {
		// if we always do PullClusterImages, we bring the docker registry, which is something
		// we may actually don't want to do here (wasted work)
		return hnd.PullClusterImages(pullOpts, cOpts.Registry, cluster)
	}
	return hnd.PullImageFromRegistry(cOpts.Registry, cluster)
}
