package filesystem

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

type RealFileSystem struct{}

type FileSystem interface {
	Open(name string) (afero.File, error)
	Glob(pattern string) ([]string, error)
	Stat(name string) (os.FileInfo, error)
}

func (fs RealFileSystem) Open(name string) (afero.File, error) {
	return os.Open(name)
}

func (fs RealFileSystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (fs RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
