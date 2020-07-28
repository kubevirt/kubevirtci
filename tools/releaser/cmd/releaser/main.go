package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	l *logrus.Logger
)

func main() {

	opts := gatherOptions(os.Args[1:])

	opts.validate()

	l = setupLogger()

	r, err := NewReleaser(opts)
	mustSucceed(err, "Could not initialize releaser")

	err = r.releaseProviders()
	mustSucceed(err, "Could not release providers")

	digestsByProvider, err := r.fetchProvidersDigest()
	mustSucceed(err, "Could not fetch provider's digest")

	buildCliOutput, err := r.buildCli(digestsByProvider)
	mustSucceed(err, buildCliOutput)

	logrus.Info(buildCliOutput)

	workingDir, err := ioutil.TempDir("/tmp", "kubevirtci-release")
	mustSucceed(err, "Could not create working dir")
	defer os.RemoveAll(workingDir)

	releaseTarballPath, err := r.buildReleaseTarball(workingDir)
	mustSucceed(err, "Could not build release tarball")

	logrus.Infof("Release tarball path: %s", releaseTarballPath)

	tag, err := r.tagRepository(filepath.Join("..", ".."))
	mustSucceed(err, "Could not tag kubevirtci repository")

	logrus.Infof("Repository HEAD tagged with %s", tag)

	err = r.createGithubRelease(tag, releaseTarballPath)
	mustSucceed(err, "Could not release kubevirtci.tar.gz at github")

	logrus.Info("Github kubevirtci release done.")
}

func mustSucceed(err error, message string) {
	if err != nil {
		logrus.WithError(err).Fatal(message)
	}
}

func setupLogger() *logrus.Logger {
	l := logrus.New()
	l.SetReportCaller(true)
	l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: time.RFC1123Z})
	l.SetLevel(logrus.TraceLevel)
	l.SetOutput(os.Stdout)
	return l
}
