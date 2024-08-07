package utils_test

import (
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

// Creates a mock directory for use within tests. This will fail the test if the directory could
// not be created
func createMockDir(fs afero.Fs, path string) {
	err := fs.MkdirAll(path, 0755)
	Expect(err).NotTo(HaveOccurred())
}

// Creates a mock file for use within tests. This will fail the test if the file could
// not be created
func createMockFile(fs afero.Fs, path string) {
	_, err := fs.Create(path)
	Expect(err).NotTo(HaveOccurred())
}

// Creates a mock file for use within tests, that contains the specified data. This will
// fail the test if the file could not be created
func createMockFileWithData(fs afero.Fs, path string, data string) {
	err := afero.WriteFile(fs, path, []byte(data), 0644)
	Expect(err).NotTo(HaveOccurred())
}
