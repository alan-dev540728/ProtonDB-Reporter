package history

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

/*
	TODO:
	change json encoder logic - next time put a full reasoning on comments, cant remember why i added this

*/

func ExtractJSONLFromArchive(archivePath, outputPath string) error {

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create JSONL directory: %w", err)
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive file: %w", err)
	}
	defer file.Close()

	gzipNewReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gzipNewReader.Close()

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create JSONL file: %w", err)
	}
	defer out.Close()

	// check Go gzip and tar documentations
	tarReader := tar.NewReader(gzipNewReader)
	foundJSON := false

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		if header.Name != "reports_piiremoved.json" {
			continue
		}

		foundJSON = true
		decoder := json.NewDecoder(tarReader)

		token, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("read JSON start: %w", err)
		}

		delimiter, ok := token.(json.Delim)
		if !ok || delimiter != '[' {
			return fmt.Errorf("expected JSON array in %s", header.Name)
		}

		for decoder.More() {
			var report map[string]any

			if err := decoder.Decode(&report); err != nil {
				return fmt.Errorf("decode report %w ", err)
			}

			if err := json.NewEncoder(out).Encode(report); err != nil {
				return fmt.Errorf("write JSONL report %w", err)
			}
		}

		endToken, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("read JSON array end: %w", err)
		}
		endDelimiter, ok := endToken.(json.Delim)
		if !ok || endDelimiter != ']' {
			return fmt.Errorf("expected end of JSON array in %s", header.Name)
		}

		break

	}

	if !foundJSON {
		return fmt.Errorf("reports_piiremoved.json not found in archive")
	}

	return nil

}
