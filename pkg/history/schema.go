package history

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

/*
 TODO:

remember some are not in the json samples

improve .db file stuff

*/

// reports table and the fields from the json
const reportsTableSchema = `
CREATE TABLE IF NOT EXISTS reports (
	id INTEGER PRIMARY KEY,
	steam_app_id TEXT NOT NULL,
	title TEXT,
	report_date TEXT,
	report_month TEXT,
	timestamp INTEGER NOT NULL,
	verdict TEXT,
	verdict_oob TEXT,
	proton_variant TEXT,
	proton_version TEXT,
	custom_proton_version TEXT,
	distro_raw TEXT,
	distro_family TEXT,
	distro_version TEXT,
	os_raw TEXT,
	cpu TEXT,
	gpu TEXT,
	gpu_driver TEXT,
	kernel TEXT,
	ram_gb INTEGER,
	steam_runtime_version TEXT,
	window_manager TEXT,
	native INTEGER NOT NULL DEFAULT 0,
	deck INTEGER NOT NULL DEFAULT 0,
	impacted_by_anticheat INTEGER NOT NULL DEFAULT 0,
	source_archive TEXT,
	source_row_number INTEGER,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

// report_date report_month index feels useless in or maybe there is a use case i forgot
const reportsIndexes = `
	 CREATE INDEX IF NOT EXISTS idx_reports_steam_app_id
	 ON reports(steam_app_id);
	 
	 CREATE INDEX IF NOT EXISTS idx_reports_report_date
	 ON reports(report_date);

	 CREATE INDEX IF NOT EXISTS idx_reports_report_month
	 ON reports(report_month);

	 CREATE INDEX IF NOT EXISTS idx_reports_proton_version
	 ON reports(proton_version);

	 CREATE INDEX IF NOT EXISTS idx_reports_distro_family
	 ON reports(distro_family);

	 CREATE INDEX IF NOT EXISTS idx_reports_distro_version
	 ON reports(distro_version);

	 CREATE INDEX IF NOT EXISTS idx_reports_gpu
	 ON reports(gpu);

	 CREATE INDEX IF NOT EXISTS idx_reports_cpu
	 ON reports(cpu);
	 `

func OpenDatabase(dbPath string) (*sql.DB, error) {
	directory := filepath.Dir(dbPath)

	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}

	if err := CreateSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil

}

func CreateSchema(db *sql.DB) error {

	if _, err := db.Exec(reportsTableSchema); err != nil {
		return fmt.Errorf("create reports table: %w", err)
	}

	if _, err := db.Exec(reportsIndexes); err != nil {
		return fmt.Errorf("create reports indexes: %w", err)
	}
	return nil
}
