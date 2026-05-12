package scraper

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func FetchMonthlyReports() ([]byte, error) {
	// https://github.com/bdefore/protondb-data/blob/master/reports/reports_jan1_2023.tar.gz
	// pattern is reports_MMM1_YYYY.tar.gz
	// https://raw.githubusercontent.com/bdefore/protondb-data/master/reports/reports_may1_2025.tar.gz

	// should skip all dates that are future in client including days or stop when there is a hit and
	// add more print to show what's happening in the fetcher
	// or just base the year and month on the current client date

	years := []int{2026, 2025, 2024}
	months := []int{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
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
			for days := 1; days <= 5; days++ {
				if found {
					break
				}
				monthDateYear = time.Date(years[year], time.Month(months[month]), days, 0, 0, 0, 0, time.UTC)
				protonDBDataRepoURL := "https://raw.githubusercontent.com/bdefore/protondb-data/master/reports/" + "reports_" + strings.ToLower(monthDateYear.Format("Jan2_2006")) + ".tar.gz"
				fullFilePath, errorDownloadFile = DownloadFile("reports_"+strings.ToLower(monthDateYear.Format("Jan2_2006"))+".tar.gz", protonDBDataRepoURL)

				if errorDownloadFile == nil {
					found = true
					break
				}

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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: failed to download %s", resp.StatusCode, url)
	}

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

func extractTarGz(fullFilePath string) ([]byte, error) {

	//var reports []models.Report

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
			/*
				err := json.NewDecoder(tarReader).Decode(&reports)
				if err != nil {
					return nil, err
				}
			*/
			jsonBytes, _ := io.ReadAll(tarReader)
			return jsonBytes, nil
		}

	}
	return nil, errors.New("reports_piiremoved.json not found in the tar.gz archive")
}
