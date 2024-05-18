package utils_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"kubevirt.io/kubevirtci/cluster-provision/centos9/vmcli/cmd/utils"
)

var _ = Describe("Environment utils", func() {
	var fs afero.Fs
	var envUtil *utils.EnvUtil

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		envUtil = utils.NewEnvUtil(fs)
	})

	Describe("Getting the node number", func() {
		Context("with an inexistent environment variable", func() {
			BeforeEach(func() {
				os.Unsetenv(utils.NodeNumKey)
			})

			It("returns the default node number", func() {
				nodeNb, err := envUtil.GetNodeNb()
				Expect(err).NotTo(HaveOccurred())
				Expect(nodeNb).To(Equal(1))
			})
		})

		Context("with an invalid environment variable", func() {
			BeforeEach(func() {
				os.Setenv(utils.NodeNumKey, "invalid")
			})

			It("fails", func() {
				_, err := envUtil.GetNodeNb()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with the environment variable set to 19", func() {
			BeforeEach(func() {
				os.Setenv(utils.NodeNumKey, "19")
			})

			It("returns the correct node number", func() {
				nodeNb, err := envUtil.GetNodeNb()
				Expect(err).NotTo(HaveOccurred())
				Expect(nodeNb).To(Equal(19))
			})
		})
	})

	Describe("Waiting for the TAP interface", func() {
		const tapNum = 1

		When("the interface does not exist", func() {
			It("times out", func() {
				err := envUtil.WaitForTap(tapNum, 1)
				Expect(err).To(MatchError(os.ErrNotExist))
			})
		})

		When("the interface exists", func() {
			BeforeEach(func() {
				createMockDir(fs, filepath.Join(utils.NetPath, fmt.Sprintf(utils.TapName, tapNum)))
			})

			It("succeeds", func() {
				err := envUtil.WaitForTap(tapNum, 1)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	DescribeTable("Checking for a rootless environment",
		func(hasFile bool, fileData string, expectedRootless bool) {
			if hasFile {
				createMockFileWithData(fs, utils.ContainerenvPath, fileData)
			}

			rootless, err := envUtil.IsRootless()
			Expect(err).NotTo(HaveOccurred())
			Expect(rootless).To(Equal(expectedRootless))
		},
		Entry("returns that the environment is not rootless when the file does not exist", false, "", false),
		Entry("returns that the environment is not rootless when the file is empty", true, "", false),
		Entry("returns that the environment is not rootless when the status is not specified in the file", true, "engine=\"podman-4.3.1\"\n", false),
		Entry("returns that the environment is not rootless when the rootless field has an invalid value", true, "engine=\"podman-4.3.1\"\nrootless=invalid\n", false),
		Entry("returns that the environment is not rootless when the rootless field has an invalid number", true, "engine=\"podman-4.3.1\"\nrootless=2\n", false),
		Entry("returns that the environment is not rootless", true, "engine=\"podman-4.3.1\"\nrootless=0\n", false),
		Entry("returns that the environment is rootless", true, "engine=\"podman-4.3.1\"\nrootless=1\n", true),
	)
})
