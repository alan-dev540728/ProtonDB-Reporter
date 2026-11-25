package history

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

/*
	TODO:
	revisit this as there seems to be an issue next time comment it so you don't forget


*/

func ValidateJSONL(inputPath string) (validCount int, invalidCount int, err error) {

	file, err := os.Open(inputPath)
	if err != nil {
		return 0, 0, fmt.Errorf("open JSONL file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())

		if len(line) == 0 {
			invalidCount++
			continue
		}

		var report map[string]json.RawMessage
		if err := json.Unmarshal(line, &report); err != nil {
			invalidCount++
			continue
		}

		if !validReportStructure(report) {
			invalidCount++
			continue
		}

		validCount++

	}

	if err := scanner.Err(); err != nil {
		return validCount, invalidCount, fmt.Errorf("scan JSONL file: %w", err)
	}

	fmt.Printf("valid rows: %d\ninvalid rows: %d\n", validCount, invalidCount)

	return validCount, invalidCount, nil

}

func validReportStructure(report map[string]json.RawMessage) bool {
	// json structure from protonDB
	var app map[string]json.RawMessage
	if !decodeObject(report["app"], &app) {
		return false
	}

	var steam map[string]json.RawMessage
	if !decodeObject(app["steam"], &steam) {
		return false
	}

	if !hasStringOrNumber(steam["appId"]) {
		return false
	}

	var title string
	if err := json.Unmarshal(app["title"], &title); err != nil || title == "" {
		return false
	}

	if !isJSONNumber(report["timestamp"]) {
		return false
	}

	var systemInfo map[string]json.RawMessage
	if !decodeObject(report["systemInfo"], &systemInfo) {
		return false
	}

	var responses map[string]json.RawMessage
	if !decodeObject(report["responses"], &responses) {
		return false
	}
	return true
}

func decodeObject(raw json.RawMessage, target *map[string]json.RawMessage) bool {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return false
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return false
	}

	return *target != nil
}

func hasStringOrNumber(raw json.RawMessage) bool {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return false
	}

	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		return stringValue != ""
	}
	return isJSONNumber(raw)
}

func isJSONNumber(raw json.RawMessage) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var number json.Number
	if err := decoder.Decode(&number); err != nil {
		return false
	}

	var extra any
	return decoder.Decode(&extra) != nil
}
