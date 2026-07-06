package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
	"github.com/AllanRaphael173554/protondb-reporter/pkg/scraper"
)

/*
Todo:
redo error handling
uniformed error logging either fmt or errors.new
fix broken forloop in fetcher
study goroutines channel and mutexes again

problem with current looping gameID to fetch is if a report does not exist instead of moving on to next gameID it stops
also only checks if input is numeric after extracting part

ram usage is around 3 to 4.9 GB

*/

func main() {
	//var outputFile bool
	var curator bool
	curator = false

	fmt.Println("Welcome to ProtonDB Reporter")
	fmt.Println("This tool fetches the latest monthly ProtonDB reports, processes them, and generates a comprehensive report for a specified game.")
	fmt.Println("This is a side project created out of my free time to practice Go Language and learn how to work with data by turning it into useful quick information. \n" +
		"All of this is possible thanks to the reports of the ProtonDB community. The data is sourced from the ProtonDB-data GitHub repository, which contains user-submitted reports on game compatibility with Proton on Linux.")
	fmt.Println("There are some limitations to this project such as accurate game rating calculation and monthly report availability.")
	fmt.Println("=================================")
	fmt.Println("Enter Game IDs to Fetch Reports:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	gameIDs := strings.Fields(strings.TrimSpace(input))
	if len(gameIDs) == 0 {
		fmt.Println("Game IDs cannot be empty.")
		return
	}

	//fmt.Println("Do you want to save the report to a file? (y/n)")
	//fmt.Scanln(&outputFile)

	fmt.Println("Fetching and processing reports for game IDs:", gameIDs)
	fetchedReports, err := scraper.FetchMonthlyReports()
	if err != nil {
		fmt.Printf("Error fetching monthly reports: %v\n", err)
		return
	}

	for _, gameID := range gameIDs {
		if _, err := strconv.Atoi(gameID); err != nil {
			//triggers after generating empty report message but it does not stop the loop and continues to next gameID
			fmt.Printf("\nSkipping invalid Game ID must be numeric: %s\n", gameID)
			continue
		}

		fmt.Printf("\nChecking reports for Game ID: %s\n", gameID)

		filteredReports, gameTitle, err := scraper.FilterReportsByGameID(fetchedReports, gameID)
		// it works on a rare case
		/*
			Fetching and processing reports for game IDs: [420]
			Fetching ProtonDB Report archive, this may take a moment...
			Found cached Archive for current month reports_jun1_2026.tar.gz
			Extracting Report Archive

			Checking reports for Game ID: 420
			Error filtering reports: unexpected end of JSON input
			No reports found for game ID 420. Please verify the game ID and try again.
		*/
		if err != nil {
			fmt.Printf("Error filtering reports: %v\n", err)
			fmt.Printf("No reports found for game ID %s. Please verify the game ID and try again.\n", gameID)
			continue
		}

		// no reports for game in archive
		if len(filteredReports) == 0 {
			fmt.Println(scraper.GenerateEmptyReportMessage(gameID))
			continue
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

		if curator {
			fmt.Println("\n===============================")
			fmt.Println("Steam Curator Micro Report")
			steamReport := scraper.GenerateSteamCuratorReport(game)
			fmt.Println(steamReport)
		}
	}

}
