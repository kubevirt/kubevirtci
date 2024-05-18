package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/spf13/afero"
)

// Util methods for virtual disks
type DiskUtil struct {
	fs afero.Fs
}

// Creates a new struct for virtual disk utils
func NewDiskUtil(fs afero.Fs) *DiskUtil {
	return &DiskUtil{
		fs,
	}
}

// Calculates the path of the next available disk image that can be generated,
// and the path of the last disk image that was generated
func (du DiskUtil) CalcNextDisk(searchDir string, forcedNextDiskPath string) (string, string, error) {
	// Get all the files at the filesystem root matching the pattern diskX.qcow2
	regex, err := regexp.Compile(`^disk(\d+).qcow2$`)
	if err != nil {
		return "", "", err
	}

	searchDir, err = filepath.Abs(searchDir)
	if err != nil {
		return "", "", err
	}

	files, err := afero.ReadDir(du.fs, searchDir)
	if err != nil {
		return "", "", err
	}

	// Find the disk with the maximum number
	lastDiskPath := "box.qcow2"
	lastDiskNb := 0

	for _, v := range files {
		fileName := v.Name()
		submatch := regex.FindStringSubmatch(fileName)

		if len(submatch) == 2 {
			i, err := strconv.Atoi(submatch[1])
			if err != nil {
				return "", "", err
			}

			if i > lastDiskNb {
				lastDiskPath = filepath.Join(searchDir, fileName)
				lastDiskNb = i
			}
		}
	}

	// Construct the name of the new disk
	nextDiskPath := fmt.Sprintf("/disk%02d.qcow2", lastDiskNb+1)

	if forcedNextDiskPath != "" {
		nextDiskPath = forcedNextDiskPath
	}

	return nextDiskPath, lastDiskPath, nil
}
