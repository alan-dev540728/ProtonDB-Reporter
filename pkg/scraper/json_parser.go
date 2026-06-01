package scraper

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
)

func FilterReportsByGameID(jsonData []byte, gameID string) ([]models.Report, string, error) {

	var rawReports []interface{}

	var filteredReports []models.Report
	var gameTitle string

	err := json.Unmarshal(jsonData, &rawReports)
	if err != nil {
		return nil, "", err
	}

	fmt.Println("Parsing Reports for " + gameID)

	for _, rawReport := range rawReports {
		report := rawReport.(map[string]interface{})
		app := report["app"].(map[string]interface{})
		steam := app["steam"].(map[string]interface{})

		// converts anything to string
		// always returns a string or empty if no matches for data type i feel like this will cause issues if protonDB changes the data type
		// this issue is present to all var that uses toString function
		appId := toString(steam["appId"])

		if appId == gameID {
			title := toString(app["title"])
			responses := report["responses"].(map[string]interface{})
			verdict := toString(responses["verdict"])
			verdictOob := toString(responses["verdictOob"])
			variant := toString(responses["variant"])
			customProtonVersion := toString(responses["customProtonVersion"])
			installs := toString(responses["installs"])
			isImpactedByAntiCheat := toString(responses["isImpactedByAntiCheat"])
			OS := toString(report["systemInfo"].(map[string]interface{})["os"])

			if variant == "older" {
				if legacyVer := toString(responses["protonVersion"]); legacyVer != "" {
					variant = legacyVer
				}
			} else if variant == "notListed" {
				if notes, ok := responses["notes"].(map[string]interface{}); ok {
					if notListedVar := toString(notes["variant"]); notListedVar != "" {
						variant = notListedVar
					}
				}
			} else if variant == "ge" {
				if customVer := toString(responses["customProtonVersion"]); customVer != "" {
					variant = customVer
				}
			}

			filteredReport := models.Report{
				AppTitle:            title,
				Installs:            installs,
				ProtonVersion:       variant,
				CustomProtonVersion: customProtonVersion,
				Verdict:             verdict,
				VerdictOutOfTheBox:  verdictOob,
				OperatingSystem:     OS,
				ImpactedByAntiCheat: stringToBool(isImpactedByAntiCheat),
			}
			filteredReports = append(filteredReports, filteredReport)

			gameTitle = title
		}
	}

	// to do about this new change
	if len(filteredReports) == 0 {
		return filteredReports, "", nil
	}

	fmt.Println("Finished parsing reports for " + gameID)
	return filteredReports, gameTitle, nil

}

func stringToBool(s string) bool {
	if s == "yes" {
		return true
	}
	return false
}

func toString(value interface{}) string {

	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatInt(int64(v), 10)
	default:
		return ""
	}

}
