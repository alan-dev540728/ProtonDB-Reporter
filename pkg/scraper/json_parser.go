package scraper

import (
	"encoding/json"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
)

func FilterReportsByGameID(jsonData []byte, gameID string) ([]models.Report, string, error) {

	var rawReports []interface{}

	var filteredReports []models.Report
	var gameTitle string
	var found bool
	err := json.Unmarshal(jsonData, &rawReports)
	if err != nil {
		return nil, "", err
	}

	for _, rawReport := range rawReports {
		report := rawReport.(map[string]interface{})
		app := report["app"].(map[string]interface{})
		steam := app["steam"].(map[string]interface{})
		appId := steam["appId"].(string)

		if appId == gameID {
			title := app["title"].(string)
			responses := report["responses"].(map[string]interface{})
			verdict := responses["verdict"].(string)
			verdictOob := responses["verdictOob"].(string)
			protonVersion := responses["protonVersion"].(string)
			customProtonVersion := responses["customProtonVersion"].(string)
			installs := responses["installs"].(string)
			isImpactedByAntiCheat := responses["isImpactedByAntiCheat"].(string)
			OS := report["systemInfo"].(map[string]interface{})["os"].(string)

			filteredReport := models.Report{
				AppTitle:            title,
				Installs:            stringToBool(installs),
				ProtonVersion:       protonVersion,
				CustomProtonVersion: customProtonVersion,
				Verdict:             verdict,
				VerdictOutOfTheBox:  verdictOob,
				OperatingSystem:     OS,
				ImpactedByAntiCheat: stringToBool(isImpactedByAntiCheat),
			}
			filteredReports = append(filteredReports, filteredReport)

			if !found {
				gameTitle = title
				found = true
			}
		}
	}

	if !found {
		return nil, "", nil
	}
	return filteredReports, gameTitle, nil

}

func stringToBool(s string) bool {
	if s == "yes" {
		return true
	}
	return false
}
