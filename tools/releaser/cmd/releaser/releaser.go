package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/google/go-github/v32/github"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	prowjobsapiv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	prowjobsclientsetv1 "k8s.io/test-infra/prow/client/clientset/versioned/typed/prowjobs/v1"
	prowconfig "k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/pjutil"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type releaser struct {
	opts        *options
	prowConfig  *prowconfig.Config
	prowJobs    prowjobsclientsetv1.ProwJobInterface
	githubToken string
}

func NewReleaser(opts *options) (*releaser, error) {

	r := &releaser{
		opts: opts,
	}

	var err error
	var restConfig *rest.Config
	if opts.kubeconfig != "" {
		restConfig, err = clientcmd.BuildConfigFromFlags("", opts.kubeconfig)
		if err != nil {
			return nil, errors.Wrapf(err, "failed instantiating K8s config from the given kubeconfig.")
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrapf(err, "failed instantiating K8s config from the in cluster config.")
		}
	}
	prowClient, err := prowjobsclientsetv1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed instantiating a Prow client from the given kubeconfig.")
	}

	err = r.initialize(prowClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed initializing releaser")
	}
	return r, nil
}

func (r *releaser) initialize(prowClient prowjobsclientsetv1.ProwV1Interface) error {
	err := r.readGithubToken()
	if err != nil {
		return errors.Wrap(err, "failed opening and reading github token at initialize")
	}

	r.prowJobs = prowClient.ProwJobs(r.opts.jobsNamespace)

	r.prowConfig, err = prowconfig.Load(r.opts.configPath, r.opts.jobConfigPath)
	if err != nil {
		return errors.Wrap(err, "failed loading prow configuration")
	}
	return nil
}

func (r *releaser) findPostsubmitConfig(name string) (prowconfig.Postsubmit, error) {
	for _, postsubmit := range r.prowConfig.JobConfig.PostsubmitsStatic[r.opts.githubOrg+"/"+r.opts.githubRepo] {
		if postsubmit.Name == name {
			return postsubmit, nil
		}
	}
	return prowconfig.Postsubmit{}, errors.Errorf("Could not find %s at postsubmit jobs configuration", name)
}

func (r *releaser) waitForProwJobCondition(name string, condition func(*prowjobsapiv1.ProwJob) (bool, error)) (prowJob *prowjobsapiv1.ProwJob, err error) {
	err = wait.PollImmediate(5*time.Second, r.opts.jobTimeout, func() (bool, error) {
		prowJob, err = r.prowJobs.Get(name, metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrapf(err, "Failed getting prowjob to for a condition")
		}
		return condition(prowJob)
	})
	return prowJob, err
}

func (r *releaser) createReleaseProviderJob(provider string) (*prowjobsapiv1.ProwJob, error) {
	providerReleaseJob := "release-" + provider
	selectedJobConfig, err := r.findPostsubmitConfig(providerReleaseJob)
	if err != nil {
		return nil, errors.Wrapf(err, "failed finding provider release job for %s", provider)
	}

	extraLabels := map[string]string{}
	extraAnnotations := map[string]string{}
	refs := prowjobsapiv1.Refs{
		Org:     r.opts.githubOrg,
		Repo:    r.opts.githubRepo,
		BaseRef: r.opts.baseRef,
		BaseSHA: r.opts.baseSha,
	}

	containerEnv := selectedJobConfig.Spec.Containers[0].Env
	containerEnv = append(containerEnv,
		corev1.EnvVar{
			Name:  "KUBEVIRTCI_ORG",
			Value: r.opts.containerOrg,
		},
		corev1.EnvVar{
			Name:  "KUBEVIRTCI_REGISTRY",
			Value: r.opts.containerRegistry,
		},
		corev1.EnvVar{
			Name:  "KUBEVIRTCI_DRYRUN",
			Value: strconv.FormatBool(r.opts.dryRun),
		},
	)
	selectedJobConfig.Spec.Containers[0].Env = containerEnv

	postSubmitJob := pjutil.NewProwJob(pjutil.PostsubmitSpec(selectedJobConfig, refs), extraLabels, extraAnnotations)

	logrus.Infof("Creating a %s prow job with name %s", providerReleaseJob, postSubmitJob.Name)
	prowJob, err := r.prowJobs.Create(&postSubmitJob)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating post submit job")
	}

	return prowJob, nil
}

func (r *releaser) releaseProviders() error {
	if len(r.opts.providers) == 0 {
		return errors.New("no provider to release specified at options")
	}
	releaseProviderJobs := []*prowjobsapiv1.ProwJob{}
	for _, provider := range r.opts.providers {
		releaseProviderJob, err := r.createReleaseProviderJob(provider)
		if err != nil {
			return errors.Wrapf(err, "failed running release provider %s job", provider)
		}
		releaseProviderJobs = append(releaseProviderJobs, releaseProviderJob)
	}

	for _, releaseProviderJob := range releaseProviderJobs {
		logrus.Infof("Waitting %+v job to complete", releaseProviderJob.Annotations["prow.k8s.io/job"])
		prowJob, err := r.waitForProwJobCondition(releaseProviderJob.Name, func(prowJobToCheck *prowjobsapiv1.ProwJob) (bool, error) {
			return prowJobToCheck.Complete(), nil
		})
		if err != nil {
			return errors.Wrap(err, "Job did not finish before timeout")
		}
		if prowJob.Status.State != prowjobsapiv1.SuccessState {
			return errors.Errorf("failed prow job with state: %s", prowJob.Status.State)
		}
		logrus.Infof("The %+v job successfully finished", releaseProviderJob.Annotations["prow.k8s.io/job"])
	}
	return nil
}

func (r *releaser) fetchProviderDigest(provider string) (string, error) {
	ref, err := name.ParseReference(filepath.Join(r.opts.containerRegistry, r.opts.containerOrg, provider))
	if err != nil {
		return "", errors.Wrapf(err, "failed parsing %s provider container URL", provider)
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", errors.Wrapf(err, "failed retrieving %s provider container", provider)
	}

	digest, err := img.Digest()
	if err != nil {
		return "", errors.Wrapf(err, "failed parsing %s provider digest", provider)

	}
	return fmt.Sprintf("%s:%s", digest.Algorithm, digest.Hex), nil
}

func (r *releaser) fetchProvidersDigest() (map[string]string, error) {
	digestByProvider := map[string]string{}
	for _, provider := range r.opts.providers {
		digest, err := r.fetchProviderDigest(provider)
		if err != nil {
			return nil, errors.Wrap(err, "failed fetching providers")
		}
		digestByProvider[provider] = digest
	}
	return digestByProvider, nil
}

func (r *releaser) buildCli(digestsByProvider map[string]string) (string, error) {
	makeArgs := []string{"-C", filepath.Join(r.opts.kubevirtciPath, "cluster-provision/gocli/"), "cli"}
	for provider, digest := range digestsByProvider {
		// Transform k8s-1.18 into something like K8S118SUFFIX
		suffixVarName := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(provider, "-", ""), ".", "")) + "SUFFIX"
		makeArgs = append(makeArgs, fmt.Sprintf("%s=\"%s\"", suffixVarName, digest))
	}
	logrus.Infof("Running 'make %s'", strings.Join(makeArgs, " "))
	cmd := exec.Command("make", makeArgs...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return string(stdoutStderr), errors.Wrap(err, "failed calling make to build cli")
	}
	return string(stdoutStderr), nil
}

func (r *releaser) buildReleaseTarball(workingDir string) (string, error) {
	tarballWorkingDir := filepath.Join(workingDir, "kubevirtci")
	tarballPath := filepath.Join(workingDir, "kubevirtci.tar.gz")

	err := copy.Copy(filepath.Join(r.opts.kubevirtciPath, "cluster-up"), tarballWorkingDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to copy 'cluster-up' directory into working dir to create release tarball")
	}

	err = copy.Copy(filepath.Join(r.opts.kubevirtciPath, "cluster-provision", "gocli", "build", "cli"), filepath.Join(tarballWorkingDir, "cli"))
	if err != nil {
		return "", errors.Wrap(err, "failed to copy 'cli' binary into working dir to create release tarball")
	}

	err = createTarball(tarballPath, tarballWorkingDir)
	if err != nil {
		return "", errors.Wrap(err, "failed creating release tarball")
	}
	return tarballPath, nil
}

func (r *releaser) tagRepository(repositoryPath string) (string, error) {
	// We instantiate a new repository targeting the given path (the .git folder)
	repo, err := git.PlainOpen(repositoryPath)
	if err != nil {
		return "", errors.Wrap(err, "failed opening kubevirtci repository")
	}

	repoHead, err := repo.Head()
	if err != nil {
		return "", errors.Wrap(err, "failed getting HEAD from kubevirtci repo")
	}

	// Use epoch as tag
	tagName := strconv.FormatInt(time.Now().Unix(), 10)

	_, err = repo.CreateTag(tagName, repoHead.Hash(), &git.CreateTagOptions{
		Message: tagName,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed tagging HEAD at kubevirtci repo")
	}

	gitCfg, err := r.composeGitConfig()
	if err != nil {
		return "", errors.Wrap(err, "failed composing git auth metod and remote URL")
	}

	cfg, err := repo.Config()
	if err != nil {
		return "", errors.Wrap(err, "failed retrieving repo config to find 'upstream' remote")
	}
	if _, remoteExists := cfg.Remotes["upstream"]; !remoteExists {
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "upstream",
			URLs: []string{gitCfg.url},
		})
		if err != nil {
			return "", errors.Wrap(err, "failed creating remote 'upstream'")
		}
	}
	po := &git.PushOptions{
		RemoteName: "upstream",
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       gitCfg.auth,
		Progress:   l.Writer(),
	}
	if r.opts.dryRun {
		logrus.Warning("dryrun: skipping pushing tags")
		return tagName, nil
	}

	err = repo.Push(po)
	if err != nil {
		return "", errors.Wrapf(err, "failed pushing kubevirtci tag to %s", gitCfg.url)
	}

	return tagName, nil
}

func (r *releaser) createGithubRelease(tag, releaseTarballPath string) error {
	body := "Follow the instruction at the tarball README to use kubevirtci"
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: r.githubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	f, err := os.Open(releaseTarballPath)
	if err != nil {
		return errors.Wrap(err, "failed opening release tarball")
	}

	if r.opts.dryRun {
		logrus.Warning("dryrun: skipping github release creation and assets upload")
		return nil
	}

	release, _, err := client.Repositories.CreateRelease(context.Background(), r.opts.githubOrg, r.opts.githubRepo, &github.RepositoryRelease{
		TagName: &tag,
		Name:    &tag,
		Body:    &body,
	})
	if err != nil {
		return errors.Wrapf(err, "failed releasing kubevirtci tarball")
	}
	_, _, err = client.Repositories.UploadReleaseAsset(context.Background(), r.opts.githubOrg, r.opts.githubRepo, *release.ID, &github.UploadOptions{Name: filepath.Base(releaseTarballPath)}, f)
	if err != nil {
		return errors.Wrap(err, "failed uploading kubevirtci release tarball to the github release")
	}
	return nil
}
