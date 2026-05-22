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
	"strconv"
	"strings"
	"time"
)

func FetchMonthlyReports() ([]byte, error) {
	// https://github.com/bdefore/protondb-data/blob/master/reports/reports_jan1_2023.tar.gz
	// pattern is reports_MMM1_YYYY.tar.gz
	// https://raw.githubusercontent.com/bdefore/protondb-data/master/reports/reports_may1_2025.tar.gz

	// should skip all dates that are future in device including days or stop when there is a hit and
	// add more print to show what's happening in the fetcher
	// or just base the year and month on the current device date

	// bug issue found if while downloading and you turn off the internet the whole CLI stops and does not update
	// however it does continue if you reconnect to the internet

	years := []int{2026}
	months := []int{5, 4, 3, 2, 1}
	var found bool
	var monthDateYear time.Time
	var fullFilePath string
	var errorDownloadFile error

	fmt.Println("Fetching ProtonDB Report archive, this may take a moment...")

	for year := range years {
		if found {
			fmt.Println("ProtonDB Report Archive Found " + strings.ToLower(monthDateYear.Format("Jan2_2006")) + " attempting to download")
			break
		}
		for month := range months {
			if found {
				break
			}
			fmt.Println("Report not found, Retrying Month " + strconv.Itoa(months[month]))
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
				} else {
					fmt.Printf("(Day %d: %v)\n", days, errorDownloadFile)
				}

			}

		}
	}

	//can trigger if year/month/day is wrong
	if !found {
		return nil, errors.New("No report archive found")
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
		fmt.Printf("error at inside downloadFile func 112: " + err.Error())

		return "", err
	}
	defer resp.Body.Close()

	// error for HTTP responses
	// is triggered by wrong URL
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: failed to download %s", resp.StatusCode, url)
	}

	// os.Mkdir errors
	// this error is directory perms
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

	fmt.Println("Extracting Report Archive")

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

			jsonBytes, _ := io.ReadAll(tarReader)
			return jsonBytes, nil
		}

	}
	return nil, errors.New("reports_piiremoved.json not found in the tar.gz archive")
}
