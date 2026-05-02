package scraper

import (
	"sort"
	"strconv"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
)

// todo implement logic and all other parts of the report struct that could be aggregated

func AggregateCommonProtonVersion(reports []models.Report) []models.VersionStats {

	if len(reports) == 0 {
		return []models.VersionStats{}
	}

	/*
		what's the best way to take all common proton version from the reports and return it as a string?

		Create a map or list to count the occurrences of each proton version
		for loop the reports and count, then find the proton version with
		the highest count and return it as a string

		I think this logic flow can be implemented to other function except rating change

	*/

	var protonVersionCount = make(map[string]int)

	for _, report := range reports {
		protonVersionCount[report.ProtonVersion]++
	}

	var protonVersions []models.VersionStats
	for version, count := range protonVersionCount {
		// sort by top to low count of proton version reported
		protonVersions = append(protonVersions, models.VersionStats{
			Version: version,
			Count:   count,
		})
	}

	sort.Slice(protonVersions, func(i, j int) bool {
		return protonVersions[i].Count > protonVersions[j].Count
	})

	if len(protonVersions) > 3 {
		return protonVersions[:3]
	}
	return protonVersions
}

func AggregateCommonLinuxDistros(reports []models.Report) []models.DistroStats {

	if len(reports) == 0 {
		return []models.DistroStats{}
	}

	distroCount := make(map[string]int)
	for _, report := range reports {
		distroCount[report.OperatingSystem]++
	}

	var linuxDistros []models.DistroStats

	for distro, count := range distroCount {
		linuxDistros = append(linuxDistros, models.DistroStats{
			Distro: distro,
			Count:  count,
		})
	}

	sort.Slice(linuxDistros, func(i, j int) bool {
		return linuxDistros[i].Count > linuxDistros[j].Count
	})

	if len(linuxDistros) > 3 {
		return linuxDistros[:3]
	}

	return linuxDistros
}

/*
func AggregateGameRatingChange(Rating string) string {

	return "No change"
}*/

/*
Possible Rating Version
90%+ platinum
70-89% Gold
40-69% Silver
<40% bronze
<10% Broken
*/

// ToDo: Implement the logic of these 2 functions and add tests for them
// Report struct
// AggregateVerdictDistribution Calculates Yes, No and Not recommended OOTB but Yes if tinker verdicts
func AggregateVerdictDistribution(reports []models.Report) string {

	if len(reports) == 0 {
		return "No change"
	}

	var verdictCountYes, verdictCountNo, verdictOOTBCountYes, verdictOOTBCountNo, totalReportCount int

	for _, report := range reports {
		totalReportCount++
		if report.Verdict == "yes" {
			verdictCountYes++

			if report.VerdictOutOfTheBox == "yes" {
				verdictOOTBCountYes++
			}

			if report.VerdictOutOfTheBox == "no" {
				verdictOOTBCountNo++
			}

		}
		if report.Verdict == "no" {
			verdictCountNo++
		}
	}

	gameRating := CalculateGameRating(verdictCountYes, verdictCountNo, verdictOOTBCountYes, verdictOOTBCountNo, totalReportCount)

	return gameRating
}

// CalculateGameRating calculates Plat, Gold, Silver, Bronze and Broken
// verdictOOTB counts to check if reports are recommending to tinker or not
func CalculateGameRating(verdictCountYes int, verdictCountNo int, verdictOOTBCountYes int, verdictOOTBCountNo int, totalReports int) string {

	if totalReports == 0 {
		return "No Reports or Broken Check API for issues"
	}
	totalVerdictRating := (verdictCountYes * 100) / totalReports
	var ootbRatingYes, ootbRatingNo, ratingTier string

	if verdictOOTBCountYes == totalReports {
		ootbRatingYes = "A total of " + strconv.Itoa(verdictOOTBCountYes) + " Working Out of the Box reports suggests that the game can be played Out of the Box"
	}

	if verdictOOTBCountNo > 0 {
		ootbRatingNo = "A total of " + strconv.Itoa(verdictOOTBCountNo) + " Not Working Out of the Box reports suggests that the game may require tinkering Proton version or some workaround to be played"
	}

	if totalVerdictRating >= 90 {
		ratingTier = "Platinum"
	} else if totalVerdictRating >= 70 {
		ratingTier = "Gold"
	} else if totalVerdictRating >= 40 {
		ratingTier = "Silver"
	} else if totalVerdictRating > 10 {
		ratingTier = "Bronze"
	} else {
		ratingTier = "Broken"
	}
	return ratingTier + " " + ootbRatingYes + " " + ootbRatingNo + " A total of " + strconv.Itoa(totalReports) + " reports"
}
