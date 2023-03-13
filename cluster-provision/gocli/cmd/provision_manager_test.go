package cmd

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/spf13/afero"
)

var fsMock afero.Fs

type MockFileSystem struct {
	fs afero.Fs
}

func (fs MockFileSystem) Open(name string) (afero.File, error) {
	return fs.fs.Open(name)
}

func (fs MockFileSystem) Glob(pattern string) ([]string, error) {
	return afero.Glob(fs.fs, pattern)
}

func (fs MockFileSystem) Stat(name string) (os.FileInfo, error) {
	return fs.fs.Stat(name)
}

var _ = BeforeSuite(func() {
	fsMock = afero.NewMemMapFs()

	dirs := []string{
		"cluster-provision/k8s/target1",
		"cluster-provision/k8s/target2",
		"cluster-up/cluster/k8s-target1",
		"cluster-up/cluster/k8s-target2",
		"cluster-up/invalid/target_none",
	}

	for _, dir := range dirs {
		err := fsMock.MkdirAll(dir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())
	}

	SetFileSystem(MockFileSystem{fsMock})
})

var _ = AfterSuite(func() {
	SetFileSystem(nil)
})

var _ = Describe("Provision Manager functionality", func() {
	Describe("processChanges", func() {
		var targets []string
		BeforeEach(func() {
			var err error
			targets, err = getTargets("cluster-provision/k8s/*")
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when rules exists in rulesDB", func() {
			It("returns expected target when matching file is Added", func() {
				rulesDB := make(map[string][]string)
				rulesDB["file1"] = []string{"target1"}

				targetToRebuild, err := processChanges(rulesDB, targets, "A\tfile1")
				Expect(err).ToNot(HaveOccurred())
				Expect(targetToRebuild).To(Equal(map[string]bool{"target1": true, "target2": false}))
			})

			It("ignore added markdown files", func() {
				rulesDB := make(map[string][]string)
				rulesDB["file1"] = []string{"target1"}

				targetToRebuild, err := processChanges(rulesDB, targets, "A\tREADME.md")
				Expect(err).ToNot(HaveOccurred())
				Expect(targetToRebuild).To(Equal(map[string]bool{"target1": false, "target2": false}))
			})

			It("fails when added file doesn't have a matching rule", func() {
				rulesDB := make(map[string][]string)
				rulesDB["file2"] = []string{"target2"}

				_, err := processChanges(rulesDB, targets, "A\tfile_should_fail")
				Expect(err.Error()).To(Equal("Errors detected: files dont have a matching rule"))
			})

			It("returns expected target when matching file is Deleted", func() {
				rulesDB := make(map[string][]string)
				rulesDB["file2"] = []string{"target2"}

				targetToRebuild, err := processChanges(rulesDB, targets, "D\tfile2")
				Expect(err).ToNot(HaveOccurred())
				Expect(targetToRebuild).To(Equal(map[string]bool{"target1": false, "target2": true}))
			})

			It("returns expected target when matching file is Renamed", func() {
				rulesDB := make(map[string][]string)
				rulesDB["file1"] = []string{"target1"}
				rulesDB["file2"] = []string{"target2"}

				targetToRebuild, err := processChanges(rulesDB, targets, "R70\tfile2\tfile1")
				Expect(err).ToNot(HaveOccurred())
				Expect(targetToRebuild).To(Equal(map[string]bool{"target1": true, "target2": false}))
			})
		})
	})

	Describe("buildRulesDBfromFile", func() {
		var targets []string
		BeforeEach(func() {
			var err error
			targets, err = getTargets("cluster-provision/k8s/*")
			Expect(err).ToNot(HaveOccurred())
		})

		DescribeTable("With valid rules.yaml file",
			func(config Config, expected map[string][]string) {
				data, err := yaml.Marshal(&config)
				Expect(err).ToNot(HaveOccurred())

				err = afero.WriteFile(fsMock, "rules.yaml", data, 0644)
				Expect(err).ToNot(HaveOccurred())

				rulesDB, err := buildRulesDBfromFile("rules.yaml", targets)
				Expect(err).ToNot(HaveOccurred())

				Expect(rulesDB).To(Equal(expected))
			},
			Entry("should create an All rule",
				Config{All: []string{"cluster-provision/k8s/*"}},
				map[string][]string{
					"cluster-provision/k8s/*": []string{"target1", "target2"},
				},
			),
			Entry("should create a None rule",
				Config{None: []string{"cluster-provision/k8s/*"}},
				map[string][]string{
					"cluster-provision/k8s/*": []string{"none"},
				},
			),
			Entry("should create a Regex rule",
				Config{Regex: []string{"cluster-provision/k8s/target[0-9]*"}},
				map[string][]string{
					"cluster-provision/k8s/target1/*": []string{"target1"},
					"cluster-provision/k8s/target2/*": []string{"target2"},
				},
			),
			Entry("should create a Regex rule on subdirectories that have 'k8s-' prefix",
				Config{Regex: []string{"cluster-up/cluster/*"}},
				map[string][]string{
					"cluster-up/cluster/k8s-target1/*": []string{"target1"},
					"cluster-up/cluster/k8s-target2/*": []string{"target2"},
				},
			),
			Entry("should create a RegexNone rule",
				Config{RegexNone: []string{"cluster-provision/k8s/target*"}},
				map[string][]string{
					"cluster-provision/k8s/target1/*": []string{"none"},
					"cluster-provision/k8s/target2/*": []string{"none"},
				},
			),
			Entry("should create an Exclude rule",
				Config{Exclude: []Exclude{Exclude{Pattern: "cluster-provision/k8s/*", Exclude: []string{"target1"}}}},
				map[string][]string{
					"cluster-provision/k8s/*": []string{"target2"},
				},
			),
			Entry("should create a Specific rule",
				Config{Specific: []Specific{Specific{Pattern: "cluster-provision/k8s/*", Targets: []string{"target1"}}}},
				map[string][]string{
					"cluster-provision/k8s/*": []string{"target1"},
				},
			),
			Entry("should create a Specific rule with multi targets",
				Config{Specific: []Specific{Specific{Pattern: "cluster-provision/k8s/*", Targets: []string{"target1", "target2"}}}},
				map[string][]string{
					"cluster-provision/k8s/*": []string{"target1", "target2"},
				},
			),
		)

		DescribeTable("With invalid rules.yaml file",
			func(config Config) {
				data, err := yaml.Marshal(&config)
				Expect(err).ToNot(HaveOccurred())

				err = afero.WriteFile(fsMock, "rules.yaml", data, 0644)
				Expect(err).ToNot(HaveOccurred())

				_, err = buildRulesDBfromFile("rules.yaml", targets)
				Expect(err).To(HaveOccurred())
			},
			Entry("should fail when invalid path used with All",
				Config{All: []string{"invalid_file"}},
			),
			Entry("should fail when invalid path used with None",
				Config{None: []string{"invalid_folder"}},
			),
			Entry("should fail when invalid path used with Regex",
				Config{Regex: []string{"invalid_folder"}},
			),
			Entry("should fail when invalid path used with RegexNone",
				Config{RegexNone: []string{"invalid_folder*"}},
			),
			Entry("should fail when invalid path used with Exclude",
				Config{Exclude: []Exclude{Exclude{Pattern: "invalid_folder/*", Exclude: []string{"target1"}}}},
			),
			Entry("should fail when invalid path used with Specific",
				Config{Specific: []Specific{Specific{Pattern: "invalid_folder/*", Targets: []string{"target1"}}}},
			),
			Entry("should fail when invalid target used with Regex",
				Config{Regex: []string{"cluster-up/invalid/target*"}},
			),
			Entry("should fail when invalid target used with Exclude",
				Config{Exclude: []Exclude{Exclude{Pattern: "cluster-provision/k8s/*", Exclude: []string{"target_na"}}}},
			),
			Entry("should fail when an Exclude rule excludes all valid targets",
				Config{Exclude: []Exclude{Exclude{Pattern: "cluster-provision/k8s/*", Exclude: []string{"target1", "target2"}}}},
			),
			Entry("should fail when invalid target used with Specific",
				Config{Specific: []Specific{Specific{Pattern: "cluster-provision/k8s/*", Targets: []string{"target_na"}}}},
			),
		)
	})

	Describe("matcher", func() {
		rulesDB := map[string][]string{
			"path/to/file":      {"rule1", "rule2"},
			"path/to/directory": {"rule2", "rule3"},
			"path/to/dir/*":     {"rule4"},
		}

		Context("when file path exists in rulesDB", func() {
			It("returns matching rules for exact file name", func() {
				matches, err := matcher(rulesDB, "path/to/file", FILE_ADDED)
				Expect(err).To(BeNil())
				Expect(matches).To(Equal([]string{"rule1", "rule2"}))
			})

			It("returns matching rules for file name in a non recursive directory", func() {
				matches, err := matcher(rulesDB, "path/to/directory/file", FILE_ADDED)
				Expect(err).To(BeNil())
				Expect(matches).To(Equal([]string{"rule2", "rule3"}))
			})

			It("returns matching rules for file name in the recursive directory", func() {
				matches, err := matcher(rulesDB, "path/to/dir/file", FILE_ADDED)
				Expect(err).To(BeNil())
				Expect(matches).To(Equal([]string{"rule4"}))
			})

			It("returns matching rules for file name in the parent directory", func() {
				matches, err := matcher(rulesDB, "path/to/dir/subdir/file", FILE_ADDED)
				Expect(err).To(BeNil())
				Expect(matches).To(Equal([]string{"rule4"}))
			})
		})

		Context("when file path does not exist in rulesDB", func() {
			It("returns an error when status is not FILE_DELETED", func() {
				_, err := matcher(rulesDB, "path/to/nonexistentfile", FILE_ADDED)
				Expect(err).NotTo(BeNil())
			})

			It("returns non error when status is FILE_DELETED", func() {
				matches, err := matcher(rulesDB, "path/to/nonexistentfile", FILE_DELETED)
				Expect(err).To(BeNil())
				Expect(matches).To(BeNil())
			})
		})
	})
})
