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

	// bug issue found if while downloading and you turn off the internet the whole CLI stops and does not update however it does continue if you reconnect to the internet

	// online and local fallback checks are added but need to add another one for previous month download if two scenarios of local previous month not existing and online current month reports not existing(fresh install and no old archive)

	now := time.Now()
	var found bool
	var monthDateYear time.Time
	var fullFilePath string
	var errorDownloadFile error
	years := now.Year()
	months := now.Month()

	fmt.Println("Fetching ProtonDB Report archive, this may take a moment...")
	for days := 1; days <= 5; days++ {
		if found {
			break
		}

		monthDateYear = time.Date(years, months, days, 0, 0, 0, 0, time.UTC)

		fileNameFormatted := "reports_" + strings.ToLower(monthDateYear.Format("Jan2_2006")) + ".tar.gz"
		localCachePath := filepath.Join("data/reports/downloads", fileNameFormatted)

		if _, err := os.Stat(localCachePath); err == nil {
			fmt.Printf("Found cached Archive for current month %s\n", fileNameFormatted)
			fullFilePath = localCachePath
			found = true
			break
		}

		protonDBDataRepoURL := "https://raw.githubusercontent.com/bdefore/protondb-data/master/reports/" + "reports_" + strings.ToLower(monthDateYear.Format("Jan2_2006")) + ".tar.gz"
		fullFilePath, errorDownloadFile = DownloadFile("reports_"+strings.ToLower(monthDateYear.Format("Jan2_2006"))+".tar.gz", protonDBDataRepoURL)

		if errorDownloadFile == nil {
			found = true
			break
		} else {
			//fmt.Printf("(Day %d: %v)\n", days, errorDownloadFile)
			fmt.Printf("Checking %s: %v\n", fileNameFormatted, errorDownloadFile)

		}

	}
	// check and review this again
	if !found {
		fmt.Println("Current Month Report Archive not Found in local or online. Checking local download for Last Month Archive")
		files, err := os.ReadDir("data/reports/downloads")
		if err == nil {
			var latestFile string
			var latestTime time.Time

			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".tar.gz") {

					info, _ := file.Info()
					if info.ModTime().After(latestTime) {
						latestTime = info.ModTime()
						latestFile = file.Name()
					}
				}
			}
			if latestFile != "" {
				fullFilePath = filepath.Join("data/reports/downloads", latestFile)
				fmt.Printf("Using local archive: %s\n", latestFile)
				found = true
			}
		}
	}
	//can trigger if year/month/day is wrong
	// would not be triggered because of !if statement above?
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
		fmt.Printf("error at inside downloadFile func 112: %s", err.Error())

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

			/*
			  var reports []models.Report
			  err := json.NewDecoder(tarReader).Decode(&reports)
			*/

			jsonBytes, _ := io.ReadAll(tarReader)
			return jsonBytes, nil
		}

	}
	return nil, errors.New("reports_piiremoved.json not found in the tar.gz archive")
}
