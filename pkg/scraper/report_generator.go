package scraper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
)

func GenerateReport(game *models.Game) string {

	verdictStr := "Game Verdict Votes: \n"
	for verdict, count := range game.VerdictDistribution {
		if verdict == "yes" {
			verdictStr += " " + verdict + ": " + strconv.Itoa(count) + " (Reports state that the game is playable or recommend to tinker) \n"
		}
		if verdict == "no" {
			verdictStr += " " + verdict + ": " + strconv.Itoa(count) + " (Reports state that the game is not playabe) \n"

		}
	}

	protonStr := "Top Common Proton Versions: \n"
	for i, versionStat := range game.TopProtonVersions {
		protonStr += " " + strconv.Itoa(i+1) + ". " + versionStat.Version + " ( " + strconv.Itoa(versionStat.Count) + " reports ) \n"
	}

	distroStr := "Top Linux Distros in Reports: \n"
	for i, distroStat := range game.TopLinuxDistros {
		distroStr += " " + strconv.Itoa(i+1) + ". " + distroStat.Distro + " ( " + strconv.Itoa(distroStat.Count) + " reports ) \n"
	}

	nativeStr := "Native Linux Reports: \n Total: " + strconv.Itoa(game.NativeReports.Count) +
		" ( " + strconv.Itoa(game.NativeReports.WorkingCount) + " working, " +
		strconv.Itoa(game.NativeReports.NotWorkingCount) + " not working ) \n"

	generatedReport := "Game Title: " + game.Title + "\n" +
		"Game SteamDeck Status: " + game.SteamDeckStatus + "\n" +
		"Official ProtonDB Rating: " + game.OfficialRating + "\n" +
		"Total ProtonDB Community Reports: " + strconv.Itoa(game.TotalReports) + "\n" +
		verdictStr +
		"Not Official Calculated Rating(Check ProtonDB Website for up-to-date Rating): " + game.CalculatedRating + "\n" +
		protonStr +
		distroStr +
		nativeStr

	return generatedReport

}

func GenerateSteamCuratorReport(game *models.Game) string {
	now := time.Now()
	var dateStr string
	var nativeStr string

	if game.NativeReports.Count == 0 {
		dateStr = now.Format("Jan 2, 2006")
		nativeStr = ""
	} else {
		dateStr = now.Format("2006-01-02")
		nativeStr = fmt.Sprintf("\nTotal Native Reports: %d, %d working, %d not working",
			game.NativeReports.Count,
			game.NativeReports.WorkingCount,
			game.NativeReports.NotWorkingCount)
	}

	return fmt.Sprintf("Game: %s\nTotal ProtonDB Community Reports: %d, you can help expanding this by submitting your own report on ProtonDB \nVerdicts: %d Yes, %d No%s \nProtonDB Rating: \nLast Updated: %s",
		game.Title,
		game.TotalReports,
		game.VerdictDistribution["yes"],
		game.VerdictDistribution["no"],
		nativeStr,
		dateStr)
}

func GenerateEmptyReportMessage(gameID string) string {
	now := time.Now()
	dateStr := now.Format("Jan 2, 2006")
	return fmt.Sprintf("No Reports found for game ID %s in this month's archive. Double check if the game ID is correct or if the game has reports in the archive.\n"+
		"Please consider making a detailed report on ProtonDB.com if you own this game "+
		"and have tested it on Linux to help the community and get an accurate report next month \n"+
		"Last Updated: %s", gameID, dateStr)
}
