package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
	"github.com/AllanRaphael173554/protondb-reporter/pkg/scraper"
)

/*
Todo:
Add more error log and handling
uniformed error logging either fmt or errors.new
fix broken forloop in fetcher
*/

func main() {
	var gameID string
	//var outputFile bool

	fmt.Println("Welcome to ProtonDB Reporter")
	fmt.Println("This tool fetches the latest monthly ProtonDB reports, processes them, and generates a comprehensive report for a specified game.")
	fmt.Println("This is a side project created out of my free time to practice Go Language and learn how to work with data by turning it into useful quick information. \n" +
		"All of this is possible thanks to the reports of the ProtonDB community. The data is sourced from the ProtonDB-data GitHub repository, which contains user-submitted reports on game compatibility with Proton on Linux.")
	fmt.Println("There are some limitations to this project such as accurate game rating calculation and monthly report availability.")
	fmt.Println("=================================")
	fmt.Println("Enter Game ID to Fetch Reports:")
	fmt.Scanln(&gameID)

	gameID = strings.TrimSpace(gameID)
	if gameID == "" {
		fmt.Println("Game ID cannot be empty.")
		return
	}

	if _, err := strconv.Atoi(gameID); err != nil {
		fmt.Printf("Game ID must be numeric")
		return
	}

	//fmt.Println("Do you want to save the report to a file? (y/n)")
	//fmt.Scanln(&outputFile)

	fmt.Println("Fetching and processing reports for game ID:", gameID)

	fetchedReports, err := scraper.FetchMonthlyReports()
	if err != nil {
		fmt.Printf("Error fetching monthly reports: %v\n", err)
		return
	}

	filteredReports, gameTitle, err := scraper.FilterReportsByGameID(fetchedReports, gameID)
	if err != nil {
		fmt.Printf("Error filtering reports: %v\n", err)
		return
	}

	verdictMap := make(map[string]int)

	for _, report := range filteredReports {
		verdictMap[report.Verdict]++
	}
	ratingStr := scraper.AggregateVerdictDistribution(filteredReports)
	nativeStats := scraper.AggregateNativeReports(filteredReports)
	topProtonVersions := scraper.AggregateCommonProtonVersion(filteredReports)
	topDistros := scraper.AggregateCommonLinuxDistros(filteredReports)

	game := &models.Game{
		ID:                  gameID,
		Title:               gameTitle,
		SteamDeckStatus:     "No Data Included",
		VerdictDistribution: verdictMap,
		CalculatedRating:    ratingStr,
		OfficialRating:      "No Data Included",
		TotalReports:        len(filteredReports),
		TopProtonVersions:   topProtonVersions,
		TopLinuxDistros:     topDistros,
		NativeReports:       nativeStats,
	}

	fmt.Println("Generating Report \n =================================")
	reportText := scraper.GenerateReport(game)
	fmt.Println(reportText)

}
