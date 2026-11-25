package history

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

/*
	TODO:
	add more exports
	need to improve this
*/

// only exports a single type of query specified at main.go
func ExportQueryToCSV(dbPath, query, outputPath string, args ...any) error {

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(cols); err != nil {
		return err
	}

	/*
		vals stores scanned values.

		ptrs stores pointers to those values.
	*/
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	for rows.Next() {

		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		record := make([]string, len(cols))
		for i := range cols {
			if vals[i] == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprint(vals[i])
			}
		}

		if err := w.Write(record); err != nil {
			return err
		}

	}

	if err := rows.Err(); err != nil {
		return err
	}

	w.Flush()
	return w.Error()

}
