package scraper

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
)

func FetchMonthlyReports() ([]models.Report, error) {
	// https://github.com/bdefore/protondb-data/blob/master/reports/reports_jan1_2023.tar.gz
	// pattern is reports_MMM1_YYYY.tar.gz
	// https://raw.githubusercontent.com/bdefore/protondb-data/master/reports/reports_may1_2025.tar.gz
	// https://gophercoding.com/download-a-file/

	years := []int{2023, 2024, 2025}
	months := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	var found bool
	var monthDateYear time.Time
	var fullFilePath string
	var errorDownloadFile error

	for year := range years {
		if found {
			break
		}
		for month := range months {
			if found {
				break
			}
			for days := 1; days <= 31; days++ {
				if found {
					break
				}
				monthDateYear = time.Date(years[year], time.Month(months[month]), days, 0, 0, 0, 0, time.UTC)
				protonDBDataRepoURL := "https://raw.githubusercontent.com/bdefore/protondb-data/master/reports/" + "reports_" + strings.ToLower(monthDateYear.Format("Jan2_2006")) + ".tar.gz"
				fullFilePath, errorDownloadFile = DownloadFile("reports_"+strings.ToLower(monthDateYear.Format("Jan2_2006"))+".tar.gz", protonDBDataRepoURL)
				if errorDownloadFile != nil {
					return nil, errorDownloadFile
				}
				found = true
				break
			}

		}
	}

	if !found {
		log.Fatal("No report file found for the specified date range.")
		return nil, errors.New("no reports found at year, month and date specified")
	}

	reports, err := extractTarGz(fullFilePath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return reports, nil
}

func DownloadFile(fileNameFormatted string, url string) (fullFilePath string, err error) {

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	//osmkdir
	errorMkDirDownloads := os.MkdirAll("data/reports/downloads", 0750)
	if errorMkDirDownloads != nil && !os.IsExist(errorMkDirDownloads) {
		log.Fatal(errorMkDirDownloads)
	}

	errorMkDirExtracted := os.MkdirAll("data/reports/extracted", 0750)
	if errorMkDirExtracted != nil && !os.IsExist(errorMkDirExtracted) {
		log.Fatal(errorMkDirExtracted)
	}

	out, err := os.Create(filepath.Join("data/reports/downloads", fileNameFormatted))
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return filepath.Join("data/reports/downloads", fileNameFormatted), err
}

func extractTarGz(fullFilePath string) ([]models.Report, error) {

	var reports []models.Report

	file, err := os.Open(fullFilePath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer file.Close()

	gzipNewReader, err := gzip.NewReader(file)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer gzipNewReader.Close()

	tarReader := tar.NewReader(gzipNewReader)

	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
			return nil, err
		}

		if header.Name == "reports_piiremoved.json" {
			err := json.NewDecoder(tarReader).Decode(&reports)
			if err != nil {
				return nil, err
			}
			break
		}

	}

	return reports, nil
}
