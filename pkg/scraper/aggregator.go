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

	var protonVersionCount = make(map[string]int)

	for _, report := range reports {
		versionToCount := report.ProtonVersion

		if versionToCount != "native" && versionToCount != "" {
			protonVersionCount[versionToCount]++
		}

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

// AggregateVerdictDistribution Calculates Yes Verdict, No Verdict and Not recommended OOB but Yes if tinker verdicts
// Game Verdict Votes
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
				//VerdictYes But OOB No is tinkering recommended
				//VerdictTinkerRecommend++
			}

		}
		if report.Verdict == "no" {
			verdictCountNo++
		}
	}

	gameRating := CalculateGameRating(verdictCountYes, verdictCountNo, verdictOOTBCountYes, verdictOOTBCountNo, totalReportCount)

	return gameRating
}

// CalculateGameRating calculates Plat, Gold, Silver, Bronze and Broken unofficially
// verdictOOTB counts to check if reports are recommending to tinker or not
func CalculateGameRating(verdictCountYes int, verdictCountNo int, verdictOOTBCountYes int, verdictOOTBCountNo int, totalReports int) string {

	if totalReports == 0 {
		return "No Reports or Broken Check API for issues"
	}
	totalVerdictRating := (verdictCountYes * 100) / totalReports
	var ootbRatingYes, ootbRatingNo, ratingTier string

	if verdictOOTBCountYes == totalReports {
		ootbRatingYes = "\n " + strconv.Itoa(verdictOOTBCountYes) + " Working Out of the Box reports suggests that the game can be played Out of the Box"
	}

	if verdictOOTBCountNo > 0 {
		ootbRatingNo = "\n " + strconv.Itoa(verdictOOTBCountNo) + " Not Working Out of the Box reports suggests that the game may require tinkering Proton version or some workaround to be played"
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
	return ratingTier + " " + ootbRatingYes + " " + ootbRatingNo + "\n A total of " + strconv.Itoa(totalReports) + " reports has been made by the community, you can help expanding this by submitting your own report on ProtonDB.com"
}

func AggregateNativeReports(reports []models.Report) models.NativeStats {

	var stats models.NativeStats

	for _, report := range reports {
		if report.ProtonVersion == "native" {
			stats.Count++
			if report.Verdict == "yes" {
				stats.WorkingCount++
			} else if report.Verdict == "no" {
				stats.NotWorkingCount++
			}
		}
	}

	return stats
}
