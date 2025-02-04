package utils

import (
	"archive/zip"
	"os"
)

// Zip создает ZIP-архив из CSV-файла
func Zip(csvPath, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	f, err := zipWriter.Create("data.csv")
	if err != nil {
		return err
	}

	data, err := os.ReadFile(csvPath)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}