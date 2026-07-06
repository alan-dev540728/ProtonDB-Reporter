# ProtonDB Reporter
A Simple CLI tool to Extract Game Specific ProtonDB monthly reports from [protondb-data](https://github.com/bdefore/protondb-data), filter by Steam game ID, aggregate verdict/Proton version/Linux distro stats per game and generate CLI text reports.

## Status

**v0.1.0-beta** — Basic functionality, still work in progress.

## Features
- Download + cache monthly reports from bdefore/protondb-data
- Filter by game ID
- Aggregate: verdict distribution, top 3 Proton versions, top 3 Linux distros, native report count per game
- Generate full text report
- Support multiple game IDs per run


## Requirements
- go1.x+
- 4GB RAM (JSON unmarshal)
- Internet for first download

## Install

```bash
git clone <repo>
cd protondb-reporter
go run .
```
No executables yet Source build only.

## Usage
go run main.go
Enter Game IDs to Fetch Reports: 70 220 420
[![Screenshot.png](https://i.postimg.cc/hjfr38rd/Screenshot.png)](https://postimg.cc/47RVY9jf)


## Known Issues
RAM usage is 3–5GB

## Features Planned
Export options:
- .txt file output for full report
- .csv export for monthly trend tracking
- .xlsx multi-sheet for distribution analysis

# Open Database License
Contains data from [protondb-data](https://github.com/bdefore/protondb-data)
ProtonDB reports data is made available under Open Database License whose full text can be found at http://opendatacommons.org/licenses/odbl/. Any rights in individual contents of the database are licensed under the Database Contents License whose text can be found http://opendatacommons.org/licenses/dbcl/