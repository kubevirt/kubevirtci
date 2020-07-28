package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type options struct {
	configPath        string
	jobConfigPath     string
	baseRef           string
	baseSha           string
	kubeconfig        string
	jobsNamespace     string
	kubevirtciPath    string
	providers         []string
	jobTimeout        time.Duration
	containerRegistry string
	containerOrg      string
	githubServer      string
	githubRepo        string
	githubOrg         string
	githubSSHKey      string
	githubToken       string
	githubUser        string
	githubHTTPSchema  string
	dryRun            bool
}

func gatherOptions(args []string) *options {
	o := &options{}

	// Used at unit test they don't need to be configurable
	o.jobTimeout = 2 * time.Hour
	o.containerRegistry = "docker.io"
	o.githubServer = "github.com"
	o.githubRepo = "kubevirtci"
	o.githubHTTPSchema = "https"

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&o.kubeconfig,
		"kubeconfig",
		"",
		"Path to kubeconfig. If empty, will try to use K8s defaults.")
	fs.StringVar(&o.jobsNamespace,
		"jobs-namespace",
		"kubevirt-prow-jobs",
		"The namespace in which Prow jobs should be created.")
	fs.StringVar(&o.configPath,
		"config-path",
		"",
		"Path to config.yaml.")
	fs.StringVar(&o.jobConfigPath,
		"job-config-path",
		"",
		"Path to prow job configs.")
	fs.StringVar(&o.baseRef,
		"base-ref",
		"master",
		"Git base ref under test")
	fs.StringVar(&o.baseSha,
		"base-sha",
		"",
		"Git base SHA under test")
	fs.StringVar(&o.kubevirtciPath,
		"kubevirtci-path",
		filepath.Join("..", ".."),
		"Path where 'cluster-up' and 'cluster-provision' is living")
	fs.StringVar(&o.containerOrg,
		"container-org",
		"kubevirtci",
		"The container registry organization")
	fs.StringVar(&o.githubOrg,
		"github-org",
		"kubevirt",
		"The github organization to use")
	fs.StringVar(&o.githubToken,
		"github-token",
		"",
		"The github token file to authenticate")
	fs.StringVar(&o.githubUser,
		"github-user",
		"",
		"The github user to authenticate")
	fs.StringVar(&o.githubSSHKey,
		"github-sshkey",
		"",
		"The github private PEM formated ssh key")
	fs.BoolVar(&o.dryRun,
		"dry-run",
		false,
		"Don't push provider, tag repo or create github release")
	providers := ""
	fs.StringVar(&providers,
		"providers",
		"k8s-1.14 k8s-1.15 k8s-1.16 k8s-1.17 k8s-1.18",
		"The github organization to use")
	fs.Parse(args)
	o.providers = strings.Split(providers, " ")
	return o
}

func (o *options) validate() {
	var errs []error
	if o.configPath == "" {
		errs = append(errs, fmt.Errorf("config-path can't be empty"))
	}
	if o.jobConfigPath == "" {
		errs = append(errs, fmt.Errorf("job-config-path can't be empty"))
	}
	if o.jobsNamespace == "" {
		errs = append(errs, fmt.Errorf("jobs-namespace can't be empty"))
	}
	if o.baseSha == "" {
		errs = append(errs, fmt.Errorf("base-sha can't be empty"))
	}
	if o.githubSSHKey == "" && (o.githubToken == "" || o.githubUser == "") {
		errs = append(errs, fmt.Errorf("missing github-sshkey or (github-token and github-user) arguments"))
	}
	if len(errs) > 0 {
		for _, err := range errs {
			logrus.WithError(err).Error("entry validation failure")
		}
		logrus.Fatalf("Arguments validation failed!")
	}
}
