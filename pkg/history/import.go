package history

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
	TODO:

	ISSUE:
	This return section has duplicate source value:
	osRaw,
	distroFamily,
	distroVersion,
	osRaw,

	Column order:
	distro_raw <- osRaw
	distro_family
	distro_version
	os_raw     <- osRaw
	so both distro_raw and os_raw currently store systemInfo.os.


That is correct only if one raw OS value intentionally serves both columns otherwise one field needs a separate source mapping

Also PopOS is not accounted for


*/

func ImportJSONLToSQLite(jsonlPath, dbPath string) (inserted int, skipped int, err error) {
	db, err := OpenDatabase(dbPath)
	if err != nil {
		return 0, 0, err
	}
	defer db.Close()

	file, err := os.Open(jsonlPath)
	if err != nil {
		return 0, 0, fmt.Errorf("open JSONL file: %w", err)
	}
	defer file.Close()

	tx, err := db.Begin()
	if err != nil {
		return 0, 0, fmt.Errorf("begin SQLite transaction: %w", err)
	}

	// delete old reports table, stacked itself with duplicate bloating the data last time
	if _, err := tx.Exec("DELETE FROM reports"); err != nil {
		_ = tx.Rollback()
		return 0, 0, fmt.Errorf("clearing reports table: %w", err)
	}

	statement, err := tx.Prepare(`
        INSERT INTO reports (
            steam_app_id,
            title,
            report_date,
            report_month,
            timestamp,
            verdict,
            verdict_oob,
            proton_variant,
            proton_version,
            custom_proton_version,
            distro_raw,
            distro_family,
            distro_version,
            os_raw,
            cpu,
            gpu,
            gpu_driver,
            kernel,
            ram_gb,
            steam_runtime_version,
            window_manager,
            native,
            deck,
            impacted_by_anticheat
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		_ = tx.Rollback()
		return 0, 0, fmt.Errorf("prepare report insert: %w", err)
	}
	defer statement.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			skipped++
			continue
		}

		report, err := decodeRawReport(line)
		if err != nil {
			skipped++
			continue
		}

		values, err := normalizeReportForDatabase(report)
		if err != nil {
			skipped++
			continue
		}

		if _, err := statement.Exec(values...); err != nil {
			skipped++
			continue
		}
		inserted++
	}

	if err := scanner.Err(); err != nil {
		_ = tx.Rollback()
		return inserted, skipped, fmt.Errorf("scan JSONL file: %w", err)
	}

	// Commit makes all tx permanent
	if err := tx.Commit(); err != nil {
		return inserted, skipped, fmt.Errorf("commit SQLite transaction: %w", err)
	}

	return inserted, skipped, nil

}

func decodeRawReport(line []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.UseNumber()

	var report map[string]any
	if err := decoder.Decode(&report); err != nil {
		return nil, err
	}
	return report, nil
}

func normalizeReportForDatabase(report map[string]any) ([]any, error) {
	app, ok := report["app"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing app object")
	}

	steam, ok := app["steam"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing app.steam object")
	}

	steamAppID := stringValue(steam["appId"])
	title := stringValue(app["title"])

	if steamAppID == "" || title == "" {
		return nil, fmt.Errorf("missing Steam App ID or title")
	}

	timestamp, err := integerValue(report["timestamp"])
	if err != nil {
		return nil, fmt.Errorf("Invalid Timestamp: %w", err)
	}

	reportTime := time.Unix(timestamp, 0).UTC()
	reportDate := reportTime.Format("2006-01-02")
	reportMonth := reportTime.Format("2006-01")

	responses, ok := report["responses"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing responses object")
	}

	systemInfo, ok := report["systemInfo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing systemInfo object")
	}

	osRaw := stringValue(systemInfo["os"])
	distroFamily, distroVersion := parseDistroInfo(osRaw)

	// Proton fields
	protonVariant := strings.ToLower(stringValue(responses["variant"]))
	customProtonVersion := stringValue(responses["customProtonVersion"])
	protonVersion := resolveProtonVersion(responses, protonVariant, customProtonVersion)

	native := protonVersion == "native" || protonVariant == "native"
	deck := strings.Contains(strings.ToLower(osRaw), "steamos")

	// issue here for duplicated OS
	return []any{
		steamAppID,
		title,
		reportDate,
		reportMonth,
		timestamp,
		stringValue(responses["verdict"]),
		stringValue(responses["verdictOob"]),
		protonVariant,
		protonVersion,
		customProtonVersion,
		osRaw,
		distroFamily,
		distroVersion,
		osRaw,
		stringValue(systemInfo["cpu"]),
		stringValue(systemInfo["gpu"]),
		stringValue(systemInfo["gpuDriver"]),
		stringValue(systemInfo["kernel"]),
		parseRAMToGB(stringValue(systemInfo["ram"])),
		stringValue(systemInfo["steamRuntimeVersion"]),
		stringValue(systemInfo["xWindowManager"]),
		boolValue(native),
		boolValue(deck),
		boolValue(stringValue(responses["isImpactedByAntiCheat"]) == "yes"),
	}, nil
}

// TODO Add documentation
func resolveProtonVersion(responses map[string]any, variant string, customProtonVersion string) string {
	switch variant {
	case "older":
		if version := stringValue(responses["protonVersion"]); version != "" {
			return version
		}
	case "ge":
		if customProtonVersion != "" {
			return customProtonVersion
		}
	case "notlisted":
		if notes, ok := responses["notes"].(map[string]any); ok {
			if version := stringValue(notes["variant"]); version != "" {
				return version
			}
		}
	}
	if version := stringValue(responses["protonVersion"]); version != "" {
		return version
	}

	if variant == "" {
		return "unknown"
	}
	return variant

}

func parseDistroInfo(value string) (family string, version string) {
	normalized := strings.TrimSpace(value)
	lower := strings.ToLower(normalized)

	// need to add/change default case for a distro not mentioned or change family = other and only mention popular ones?
	switch {
	case strings.HasPrefix(lower, "ubuntu"):
		family = "Ubuntu"
	case strings.HasPrefix(lower, "linux mint"):
		family = "Linux Mint"
	case strings.HasPrefix(lower, "fedora"):
		family = "Fedora"
	case strings.HasPrefix(lower, "arch"):
		family = "Arch"
	case strings.HasPrefix(lower, "manjaro"):
		family = "Manjaro"
	case strings.HasPrefix(lower, "debian"):
		family = "Debian"
	case strings.HasPrefix(lower, "steamos"):
		family = "SteamOS"
	case strings.HasPrefix(lower, "bazzite"):
		family = "Bazzite"
	case strings.HasPrefix(lower, "cachyos"):
		family = "CachyOS"
	case strings.HasPrefix(lower, "nobara"):
		family = "Nobara"
	default:
		family = "Unknown"
	}

	fields := strings.Fields(normalized)
	for _, field := range fields {
		field = strings.Trim(field, "()[],")

		if strings.Count(field, ".") == 1 {
			parts := strings.Split(field, ".")
			if isDigits(parts[0]) && isDigits(parts[1]) {
				return family, field
			}
		}
	}

	return family, ""

}

func parseRAMToGB(value string) int {
	fields := strings.Fields(strings.ToUpper(value))
	if len(fields) == 0 {
		return 0
	}

	number, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0
	}

	return number

}

func stringValue(value any) string {
	switch value := value.(type) {
	case string:
		return strings.TrimSpace(value)
	case json.Number:
		return value.String()
	case float64:
		return strconv.FormatInt(int64(value), 10)
	default:
		return ""
	}
}

func integerValue(value any) (int64, error) {
	switch value := value.(type) {
	case json.Number:
		return value.Int64()
	case float64:
		return int64(value), nil
	case string:
		return strconv.ParseInt(value, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported number type: %T", value)
	}
}

func boolValue(value bool) int {
	if value {
		return 1
	}
	return 0
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
