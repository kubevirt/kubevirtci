package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func createTarball(tarballFilePath string, baseDirPath string) error {

	file, err := os.Create(tarballFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed creating tarball file '%s'", tarballFilePath)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(baseDirPath,
		func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			err = addFileToTarWriter(baseDirPath, filePath, info, tarWriter)
			if err != nil {
				return errors.Wrapf(err, "failed adding file '%s' to tarball", filePath)
			}

			return nil
		})
	if err != nil {
		return errors.Wrap(err, "failed 'walking' files to add them to tarball")
	}
	return nil
}

func addFileToTarWriter(baseDirPath, filePath string, info os.FileInfo, tarWriter *tar.Writer) error {
	// Compose tar header from file info
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return errors.Wrap(err, "failed creating tar header from file stats")
	}

	// Store files in tar as relative paths
	relFilePath, err := filepath.Rel(baseDirPath, filePath)
	if err != nil {
		return errors.Wrap(err, "failed composing relative path from file to store at tarball")
	}
	header.Name = relFilePath

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return errors.Wrapf(err, "failed writting header for file '%s'", filePath)
	}

	// This is a dir nothing more to do
	if info.IsDir() {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed opening file '%s'", filePath)
	}
	defer file.Close()

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return errors.Wrapf(err, "failed to copy file '%s' data to the tarball", filePath)
	}

	return nil
}

func extractTarball(tarballPath, extractDir string) error {

	tarballFile, err := os.Open(tarballPath)
	if err != nil {
		return errors.Wrapf(err, "failed openening tarball %s", tarballPath)
	}

	uncompressedStream, err := gzip.NewReader(tarballFile)
	if err != nil {
		return errors.Wrap(err, "failed creating new gzip reader to uncompress tarball")
	}

	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		return errors.Wrap(err, "failed creating directory where tarball will be extracted")
	}

	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed reading next file from tarball")
		}

		// determine proper file path info
		fileInfo := header.FileInfo()
		fileName := filepath.Join(extractDir, header.Name)

		if header.Typeflag == tar.TypeDir {
			err := os.MkdirAll(fileName, 0755)
			if err != nil {
				return errors.Wrap(err, "failed creating directory from tarball")
			}
			continue
		}

		file, err := os.OpenFile(
			fileName,
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			fileInfo.Mode().Perm(),
		)
		if err != nil {
			return errors.Wrap(err, "failed creating file from tarball")
		}
		defer file.Close()
		n, err := io.Copy(file, tarReader)
		if err != nil {
			return errors.Wrap(err, "failed to copy file from tarball")
		}
		if n != fileInfo.Size() {
			return errors.Errorf("wrote %d, want %d", n, fileInfo.Size())
		}
	}
	return nil
}
