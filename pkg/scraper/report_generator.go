package scraper

import (
	"strconv"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
)

func GenerateReport(game *models.Game) string {

	// todo better formatting

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
