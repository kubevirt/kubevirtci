package filesystem

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var fileSystem FileSystem = RealFileSystem{}

type FileSystem interface {
	Open(name string) (afero.File, error)
	Glob(pattern string) ([]string, error)
	Stat(name string) (os.FileInfo, error)
}

// RealFileSystem
type RealFileSystem struct{}

func (fs RealFileSystem) Open(name string) (afero.File, error) {
	return os.Open(name)
}

func (fs RealFileSystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (fs RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// MockFileSystem
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

// General

func SetMockFileSystem() {
	fileSystem = MockFileSystem{fs: afero.NewMemMapFs()}
}

func SetRealFileSystem() {
	fileSystem = RealFileSystem{}
}

func GetFileSystem() FileSystem {
	return fileSystem
}

func GetFs() afero.Fs {
	if mockFs, ok := fileSystem.(MockFileSystem); ok {
		return mockFs.fs
	} else {
		return afero.NewOsFs()
	}
}

func GlobDirectories(path string) ([]string, error) {
	files, err := fileSystem.Glob(path)
	if err != nil {
		return nil, err
	}

	var directories []string
	for _, candid := range files {
		f, err := fileSystem.Stat(candid)
		if err != nil {
			return nil, err
		}
		if f.IsDir() {
			directories = append(directories, candid)
		}
	}

	return directories, nil
}
