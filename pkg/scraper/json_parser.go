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
	var found bool
	err := json.Unmarshal(jsonData, &rawReports)
	if err != nil {
		return nil, "", err
	}

	for _, rawReport := range rawReports {
		report := rawReport.(map[string]interface{})
		app := report["app"].(map[string]interface{})
		steam := app["steam"].(map[string]interface{})

		appId, err := toStringAppId(steam["appId"])
		if err != nil {
			return nil, "", fmt.Errorf("error converting appId to string: %v", err)
		}

		if appId == gameID {
			title := toStringField(app["title"])
			responses := report["responses"].(map[string]interface{})
			verdict := toStringField(responses["verdict"])
			verdictOob := toStringField(responses["verdictOob"])
			variant := toStringField(responses["variant"])
			customProtonVersion := toStringField(responses["customProtonVersion"])
			installs := toStringField(responses["installs"])
			isImpactedByAntiCheat := toStringField(responses["isImpactedByAntiCheat"])
			OS := toStringField(report["systemInfo"].(map[string]interface{})["os"])

			if variant == "older" {
				if legacyVer := toStringField(responses["protonVersion"]); legacyVer != "" {
					variant = legacyVer
				}
			} else if variant == "notListed" {
				if notes, ok := responses["notes"].(map[string]interface{}); ok {
					if notListedVar := toStringField(notes["variant"]); notListedVar != "" {
						variant = notListedVar
					}
				}
			} else if variant == "ge" {
				if customVer := toStringField(responses["customProtonVersion"]); customVer != "" {
					variant = customVer
				}
			}

			filteredReport := models.Report{
				AppTitle:            title,
				Installs:            stringToBool(installs),
				ProtonVersion:       variant,
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

func toStringField(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatInt(int64(v), 10)
	default:
		return ""
	}
}

func toStringAppId(value interface{}) (string, error) {
	if value == nil {
		return "", fmt.Errorf("appId is nil")
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatInt(int64(v), 10), nil
	default:
		return "", fmt.Errorf("unexpected type for appId: %T", value)
	}
}
