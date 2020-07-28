package main

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	projectInfraPath string
)

var _ = BeforeSuite(func() {

	var err error
	projectInfraPath, err = ioutil.TempDir("/tmp", "project-infra")
	Expect(err).To(Succeed(), "should succeed creating tmp dir for project-infra")

	By("Cloning project-infra")
	_, err = git.PlainClone(projectInfraPath, false, &git.CloneOptions{
		URL:           "https://github.com/qinqon/project-infra",
		ReferenceName: plumbing.NewBranchReferenceName("kubevirtci-release"),
		SingleBranch:  true,
	})
	Expect(err).To(Succeed(), "should succeed cloning project-infra")
})

var _ = AfterSuite(func() {
	os.RemoveAll(projectInfraPath) // clean up
})

func TestReleaser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Releaser Test Suite")
}
