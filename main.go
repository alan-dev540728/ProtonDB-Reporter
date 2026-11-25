package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/history"
	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
	"github.com/AllanRaphael173554/protondb-reporter/pkg/scraper"
)

/*
Todo:
redo error handlings
uniformed error logging either fmt or errors.new
fix broken forloop in fetcher
study goroutines channel and mutexes again

problem with current looping gameID to fetch is if a report does not exist instead of moving on to next gameID it stops
also only checks if input is numeric after extracting part

RAM usage seems fixed but it still download fetch everytime
basically just hovering around 2GB of usage and not going up or down but there is always a
download and upload activity whenever it moves to a new gameID

Redo Todos and notes and comments
add gosec and CI
fix paths
make permanent fix to choose normal or curator report and db export
add dynamic monthDay format to tar.gz

update reports to move "you can expand" to the end of the report instead of start
update reports to add extra new lines between sentences for easy copy paste

add proper json structure documentation to this project

change downloads location to protondb/downloads
change output to reports so for game specific reviews reports/curator_report reports/full_report reports/something.csv
*/

func main() {
	var saveToFile bool
	var allReports strings.Builder

	fmt.Println("Welcome to ProtonDB Reporter")
	fmt.Println("This tool fetches the latest monthly ProtonDB reports, processes them, and generates a comprehensive report for a specified game.")
	fmt.Println("This is a side project created out of my free time to practice Go Language and learn how to work with data by turning it into useful quick information. \n" +
		"All of this is possible thanks to the reports of the ProtonDB community. The data is sourced from the ProtonDB-data GitHub repository, which contains user-submitted reports on game compatibility with Proton on Linux.")
	fmt.Println("There are some limitations to this project such as accurate game rating calculation and monthly report availability.")
	fmt.Println("=================================")

	gameIDs := readGameIDs()

	// Save to File prompt
	saveToFile = saveToFilePrompt()

	fmt.Println("Fetching and processing reports for game IDs:", gameIDs)

	if err := runHistoryPipeline(); err != nil {
		fmt.Println("History pipeline failed:", err)
		return
	}
	for _, gameID := range gameIDs {

		if !isGameIDNumeric(gameID) {
			fmt.Printf("\nSkipping invalid Game ID must be numeric: %s\n", gameID)
			continue
		}

		reportText, err := processGameIDs(gameID)
		if err != nil {
			if saveToFile {
				allReports.WriteString("Game ID: " + gameID + "\n")
				allReports.WriteString(scraper.GenerateEmptyReportMessage(gameID) + "\n")
				allReports.WriteString("\n\n==============================\n\n")
			} else {
				fmt.Println(err.Error())
			}
			continue
		}
		if saveToFile {
			allReports.WriteString("Game ID: " + gameID + "\n")
			allReports.WriteString(reportText + "\n")
			allReports.WriteString("\n\n==============================\n\n")
		} else {
			fmt.Println(reportText)
		}

	}

	if saveToFile {

		err := os.MkdirAll("data/reports/output", 0755)
		if err != nil {
			fmt.Println("Error creating reports directory:", err)
		}

		err = os.WriteFile("data/reports/output/all_reports.txt", []byte(allReports.String()), 0644)
		if err != nil {
			fmt.Println("Writing to file failed:", err)
			return
		}
		fmt.Println("All reports saved to data/reports/output/all_reports.txt")
	}

}

func readGameIDs() []string {
	// Steam Game ID input
	fmt.Println("Enter Game IDs to Fetch Reports:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	gameIDs := strings.Fields(strings.TrimSpace(input))
	if len(gameIDs) == 0 {
		fmt.Println("Game IDs cannot be empty.")
		// need to exit here or prompt again for input
		os.Exit(1)
	}
	return gameIDs
}

func saveToFilePrompt() bool {
	var save bool

	fmt.Println("Save reports to file? (y/n)")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "n" {
		fmt.Println("Invalid input Defaulting to 'n'.")
		save = false
	} else {
		if answer == "y" {
			save = true
		} else {
			save = false
		}
	}
	return save
}

func isGameIDNumeric(gameID string) bool {
	if _, err := strconv.Atoi(gameID); err != nil {
		// triggers on incorrect Game ID with string text
		// triggers after generating empty report message but it does not stop the loop and continues to next gameID
		// is not written for all_reports.txt
		return false
	}
	return true
}

func buildGameModel(gameID, gameTitle string, reports []models.Report) *models.Game {
	verdictMap := make(map[string]int)

	for _, report := range reports {
		verdictMap[report.Verdict]++
	}

	ratingStr := scraper.AggregateVerdictDistribution(reports)
	nativeStats := scraper.AggregateNativeReports(reports)
	topProtonVersions := scraper.AggregateCommonProtonVersion(reports)
	topDistros := scraper.AggregateCommonLinuxDistros(reports)

	return &models.Game{
		ID:                  gameID,
		Title:               gameTitle,
		SteamDeckStatus:     "No Data Included",
		VerdictDistribution: verdictMap,
		CalculatedRating:    ratingStr,
		OfficialRating:      "No Data Included",
		TotalReports:        len(reports),
		TopProtonVersions:   topProtonVersions,
		TopLinuxDistros:     topDistros,
		NativeReports:       nativeStats,
	}

}

func generateReportText(game *models.Game, curator bool) string {
	if curator {
		fmt.Println("\n===============================")
		fmt.Println("Steam Curator Micro Report")
		curatorReportText := scraper.GenerateSteamCuratorReport(game)
		fmt.Println("DEBUG 218:" + curatorReportText)
		return curatorReportText
	} else {
		fmt.Println("Generating Report \n =================================")
		normalReportText := scraper.GenerateReport(game)
		fmt.Println(normalReportText)
		return normalReportText
	}

}

func processGameIDs(gameID string) (string, error) {
	var curator bool
	curator = false

	fmt.Printf("\nChecking reports for Game ID: %s\n", gameID)
	filteredReports, gameTitle, err := scraper.FetchMonthlyReports(gameID)

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
	if err != nil || len(filteredReports) == 0 {
		// Error filtering reports: <nil>
		fmt.Printf("Error filtering reports: %v\n", err)
		fmt.Printf("No reports found for game ID %s. Please verify the game ID and try again.\n", gameID)

	}

	// no reports for game in archive
	if len(filteredReports) == 0 {
		return "", fmt.Errorf(scraper.GenerateEmptyReportMessage(gameID))
	}

	game := buildGameModel(gameID, gameTitle, filteredReports)

	return generateReportText(game, curator), nil

}

// needs to be updated to select between curator/normal report or db and csv export
// runHistoryPipeline runs before generate report
func runHistoryPipeline() error {
	// also need to make the path of monthDay and year dynamic
	archivePath := "data/reports/downloads/reports_nov1_2026.tar.gz"
	jsonlPath := "data/protondb/jsonl/reports_nov1_2026.jsonl"
	dbPath := "data/protondb/database/protondb_history.db"
	csvPath := "data/reports/output/total_reports.csv"

	// extract
	if err := history.ExtractJSONLFromArchive(archivePath, jsonlPath); err != nil {
		return err
	}

	// validate
	valid, invalid, err := history.ValidateJSONL(jsonlPath)
	if err != nil {
		return err
	}

	fmt.Printf("valid: %d, invalid: %d\n", valid, invalid)

	// import
	inserted, skipped, err := history.ImportJSONLToSQLite(jsonlPath, dbPath)
	if err != nil {
		return err
	}

	fmt.Printf("inserted: %d, skipped: %d\n", inserted, skipped)

	// Query Export to CSV
	// database path, query, CSV output path, args
	// Hardcoded
	return history.ExportQueryToCSV(dbPath, history.QueryGameProtonVersions, csvPath, "1745790")

}
