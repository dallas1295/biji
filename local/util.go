package local

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func createZip(source string, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("error creating target directory %w", err)
	}

	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Base(path)
		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		if info.Size() == 0 {
			file.Close()
			return nil
		}

		_, err = io.Copy(writer, file)
		file.Close()

		return err
	})
}
