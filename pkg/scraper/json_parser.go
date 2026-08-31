package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/AllanRaphael173554/protondb-reporter/pkg/models"
)

/* json unmarshal vs json.Decoder
need to rework this in the future
*/

func FilterReportsByGameID(reader io.Reader, gameID string) ([]models.Report, string, error) {

	decoder := json.NewDecoder(reader)

	token, err := decoder.Token()
	if err != nil {
		return nil, "", err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return nil, "", fmt.Errorf("expected JSON array, got %v", token)
	}

	var filteredReports []models.Report
	var gameTitle string

	for decoder.More() {
		var rawReport map[string]interface{}

		if err := decoder.Decode(&rawReport); err != nil {
			return nil, "", fmt.Errorf("decode report: %w", err)
		}

		report, matches := parseReport(rawReport, gameID)
		if !matches {
			continue
		}

		filteredReports = append(filteredReports, report)
		if gameTitle == "" {
			gameTitle = report.AppTitle
		}
	}

	if _, err := decoder.Token(); err != nil {
		return nil, "", err
	}
	return filteredReports, gameTitle, nil

	/* this needs to be reworked

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
	*/
}

func parseReport(rawReport map[string]interface{}, gameID string) (models.Report, bool) {

	app, ok := rawReport["app"].(map[string]interface{})
	if !ok {
		return models.Report{}, false
	}

	steam, ok := app["steam"].(map[string]interface{})
	if !ok {
		return models.Report{}, false
	}

	if toString(steam["appId"]) != gameID {
		return models.Report{}, false
	}

	responses, ok := rawReport["responses"].(map[string]interface{})
	if !ok {
		return models.Report{}, false
	}

	systemInfo, _ := rawReport["systemInfo"].(map[string]interface{})

	title := toString(app["title"])
	variant := resolveVariant(responses)

	return models.Report{
		AppTitle:            title,
		Installs:            toString(responses["installs"]),
		ProtonVersion:       variant,
		CustomProtonVersion: toString(responses["customProtonVersion"]),
		Verdict:             toString(responses["verdict"]),
		VerdictOutOfTheBox:  toString(responses["verdictOob"]),
		OperatingSystem:     toString(systemInfo["os"]),
		ImpactedByAntiCheat: stringToBool(toString(responses["isImpactedByAntiCheat"])),
	}, true

}

func resolveVariant(responses map[string]interface{}) string {
	variant := toString(responses["variant"])

	switch variant {
	case "older":
		if value := toString(responses["protonVersion"]); value != "" {
			return value
		}
	case "notListed":
		notes, _ := responses["notes"].(map[string]interface{})
		if value := toString(notes["variant"]); value != "" {
			return value
		}
	case "ge":
		if value := toString(responses["customProtonVersion"]); value != "" {
			return value
		}
	}

	return variant
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
