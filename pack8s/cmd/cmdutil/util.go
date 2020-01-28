package cmdutil

import (
	"context"
	"os"
	"strconv"

	logger "github.com/apsdehal/go-logger"
	isatty "github.com/mattn/go-isatty"

	"github.com/spf13/cobra"

	"github.com/fromanirh/pack8s/internal/pkg/podman"
)

type CommonOpts struct {
	Prefix       string
	PodmanSocket string
	Verbose      int
	Registry     string
	IsTTY        bool
	Color        bool
}

func AddCommonOpts(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringP("prefix", "p", "kubevirt", "Prefix to identify containers")
	rootCmd.PersistentFlags().StringP("podman-socket", "s", podman.DefaultSocket, "Path to podman-socket")
	rootCmd.PersistentFlags().IntP("verbose", "v", 3, "verbosiness level [1,5)")
	rootCmd.PersistentFlags().StringP("container-registry", "R", "docker.io", "Registry to pull cluster images from")
}

func GetCommonOpts(cmd *cobra.Command) (CommonOpts, error) {
	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return CommonOpts{}, err
	}
	podmanSocket, err := cmd.Flags().GetString("podman-socket")
	if err != nil {
		return CommonOpts{}, err
	}
	verbose, err := cmd.Flags().GetInt("verbose")
	if err != nil {
		return CommonOpts{}, err
	}
	registry, err := cmd.Flags().GetString("container-registry")
	if err != nil {
		return CommonOpts{}, err
	}

	if val, ok := os.LookupEnv("PACK8S_VERBOSE"); ok {
		if v, err := strconv.Atoi(val); err == nil {
			verbose = v
		}
	}

	color := false
	if val, ok := os.LookupEnv("PACK8S_COLORS"); ok {
		color = isTruish(val)
	}

	return CommonOpts{
		Prefix:       prefix,
		PodmanSocket: podmanSocket,
		Verbose:      verbose,
		IsTTY:        isatty.IsTerminal(os.Stderr.Fd()),
		Color:        color,
		Registry:     registry,
	}, nil
}

func (co CommonOpts) GetLogger() *logger.Logger {
	return NewLogger(co.Verbose, co.Color, co.IsTTY)
}

func (co CommonOpts) GetHandle() (*podman.Handle, *logger.Logger, error) {
	ctx := context.Background()
	log := co.GetLogger()
	hnd, err := podman.NewHandle(ctx, co.PodmanSocket, log)
	return hnd, log, err
}

func NewLogger(lev int, color, tty bool) *logger.Logger {
	// go-logger doesn't accept bool, uses ints as bools (?)
	colored := 0
	if color {
		colored = 1
	}
	log, err := logger.New("pack8s", colored, toLogLevel(lev))
	if err != nil {
		panic(err)
	}
	if tty { // interactive
		log.SetFormat("%{message}")
	} else {
		log.SetFormat("%{time} %{message}")
	}
	log.Debugf("logger level: %v", lev)
	return log
}

func toLogLevel(lev int) logger.LogLevel {
	switch {
	// ALWAYS emit crit messages
	case lev <= 1:
		return logger.ErrorLevel
	case lev == 2:
		return logger.WarningLevel
	case lev == 3:
		return logger.NoticeLevel
	case lev == 4:
		return logger.InfoLevel
	case lev >= 5:
		return logger.DebugLevel
	}
	// should never be reached
	return logger.InfoLevel
}

func isTruish(v string) bool {
	switch v {
	case "1":
		return true
	case "Y", "y":
		return true
	case "YES", "yes":
		return true
	}
	return false
}
