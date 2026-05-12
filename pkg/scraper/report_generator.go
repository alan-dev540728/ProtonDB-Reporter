package scraper

import (
	"strconv"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
)

func GenerateReport(game *models.Game) string {

	// todo better formatting

	verdictStr := "Game Verdict Votes: \n"
	for verdict, count := range game.VerdictDistribution {
		verdictStr += " " + verdict + ": " + strconv.Itoa(count) + "\n"
	}

	protonStr := "Top Common Proton Versions: \n"
	for i, versionStat := range game.TopProtonVersions {
		protonStr += " " + strconv.Itoa(i+1) + ". " + versionStat.Version + " ( " + strconv.Itoa(versionStat.Count) + " reports ) \n"
	}

	distroStr := "Top Linux Distros in Reports: \n"
	for i, distroStat := range game.TopLinuxDistros {
		distroStr += " " + strconv.Itoa(i+1) + ". " + distroStat.Distro + " ( " + strconv.Itoa(distroStat.Count) + " reports ) \n"
	}

	generatedReport := "Game Title: " + game.Title + "\n" +
		"Game SteamDeck Status: " + game.SteamDeckStatus + "\n" +
		"Total ProtonDB Community Reports: " + strconv.Itoa(game.TotalReports) + "\n" +
		verdictStr +
		"Calculated Rating: " + game.CalculatedRating + "\n" +
		"Official ProtonDB Rating: " + game.OfficialRating + "\n" +
		protonStr +
		distroStr

	return generatedReport

}
