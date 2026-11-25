package history

// starts from reports table in protondb_history.db
/*
   TODO: Update query names to prevent confusion again
*/

// counts report rows by the month in each report's timestamp.
// cumulative_reports is the running sum through that month neither count deduplicates users or games
const QueryReportsByMonth = `
WITH monthly AS (
    SELECT
        report_month,
        COUNT(*) AS reports_in_month,
        MIN(report_date) AS first_report_date,
        MAX(report_date) AS latest_report_date
    FROM reports
    GROUP BY report_month
)
SELECT
    report_month,
    reports_in_month,
    SUM(reports_in_month) OVER (
        ORDER BY report_month
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS cumulative_reports,
    first_report_date,
    latest_report_date
FROM monthly
ORDER BY report_month;
	`

// games_with_reports: distinct games with reports that month, including games seen before.
// new_games_first_seen: games whose earliest report in the database is that month.
/*
   games_with_reports = 1,715: 1,715 distinct games had at least one report dated that month. Includes previously reported games and first-seen games.
   new_games_first_seen = 580: 580 games had their earliest report in the loaded database that month. These are included in the 1,715
   game without reports not counted
*/
const QueryNewGamesFirstSeenPerMonth = `
WITH monthly_activity AS (
    SELECT
        report_month,
        COUNT(DISTINCT steam_app_id) AS games_with_reports
    FROM reports
    GROUP BY report_month
),
first_seen AS (
    SELECT
        steam_app_id,
        MIN(report_month) AS first_month
    FROM reports
    GROUP BY steam_app_id
),
monthly_new AS (
    SELECT
        first_month AS report_month,
        COUNT(*) AS new_games_first_seen
    FROM first_seen
    GROUP BY first_month
)
SELECT
    activity.report_month,
    activity.games_with_reports,
    COALESCE(new.new_games_first_seen, 0) AS new_games_first_seen,
    SUM(COALESCE(new.new_games_first_seen, 0)) OVER (
        ORDER BY activity.report_month
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS cumulative_unique_games
FROM monthly_activity AS activity
LEFT JOIN monthly_new AS new USING (report_month)
ORDER BY activity.report_month;
`

// same as above but unique only
const QueryUniqueGamesTotal = `
WITH first_seen AS (
    SELECT
        steam_app_id,
        MIN(report_month) AS first_month
    FROM reports
    GROUP BY steam_app_id
),
monthly AS (
    SELECT
        first_month AS report_month,
        COUNT(*) AS new_unique_games
    FROM first_seen
    GROUP BY first_month
)
SELECT
    report_month,
    new_unique_games,
    SUM(new_unique_games) OVER (
        ORDER BY report_month
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS cumulative_unique_games
FROM monthly
ORDER BY report_month;
`

// per month report of each proton version used in reports and how many times it was used in reports
// looking messy maybe add option to make it by year only or make specific month option
// measures volume not unique users
const QueryProtonVersionsByMonth = `
SELECT
    report_month,
    proton_version,
    COUNT(*) AS total_reports
FROM reports
GROUP BY report_month, proton_version
ORDER BY report_month, total_reports DESC;
`

// i am thinking of doing a QueryGlobalDistrosByMonth or Year to make a full trend instead of just by months trend as each reports are individual and not indicative of overall total
// meaning that a user with Linux Mint can send the same report for the same game multiple times(maybe normal report vs modded proton report) or different games making the count bloated as it will only count for each existing and not unique
// cannot make unique PII is removed
// measures volume not unique users
const QueryDistrosByMonth = `
SELECT
    report_month,
    distro_family,
    COUNT(*) AS total_mentions
FROM reports
GROUP BY report_month, distro_family
ORDER BY report_month, total_mentions DESC;
`

// measures volume not unique users
// should i even add a per year total of proton versions used in reports?
// probably change to variant instead of ProtonVersions or make new query
/*
   TODO:
   find out the differnce again between proton_version
   official, Default and experimental
    there should be a comment/note about this somewhere
*/
/*
Year   Proton version   total_reports
2026        Default         100
        Experimental    200
        GE              300
*/
const QueryGlobalProtonVersions = `
SELECT
    proton_version,
    COUNT(*) AS total_reports
FROM reports
GROUP BY proton_version
ORDER BY total_reports DESC;
`

// basically trend check? for distro users seeing which distro is most used by protondb users from number of reports
// note PopOS not included yet
// measures volume not unique users
const QueryGlobalDistros = `
SELECT
    distro_family,
    COUNT(*) AS total_mentions
FROM reports
GROUP BY distro_family
ORDER BY total_mentions DESC;
`

// common distro but specific for games only
// may have bloated data as some users may send multiple reports (valid since some user send reports with GE or DWProton but same distro)
// issue with separating distros just because of distro_version | remove soon
const QueryGameDistroCounts = `
SELECT
    distro_family,
    distro_version,
    COUNT(*) AS total_reports
FROM reports
WHERE steam_app_id = ?
GROUP BY distro_family, distro_version
ORDER BY total_reports DESC;
`

// common total proton versions but specific game only separated by month and versions
const QueryGameProtonVersions = `
SELECT
    report_month,
    proton_version,
    COUNT(*) AS total_reports
FROM reports
WHERE steam_app_id = ?
GROUP BY report_month, proton_version
ORDER BY report_month, total_reports DESC;
`

// one game grouped by month and GPU
/*
   TODO: improve this for better grouping | i am sure these are the same GPUs
   sample 2025-11
   AMD Radeon RX 9070 XT (radeonsi, gfx1201, LLVM 21.1.5, DRM 3.64, 6.17.7-zen1-2-zen)
   AMD Radeon RX 9070 XT (radeonsi, gfx1201, LLVM 20.1.8, DRM 3.64, 6.17.7-200.nobara.fc42.x86_64)
   AMD Radeon RX 9070 XT (radeonsi, gfx1201, ACO, DRM 3.64, 6.17.8-2-cachyos)
   AMD Radeon RX 7900 XTX (radeonsi, navi31, LLVM 19.1.7, DRM 3.61, 6.12.41+deb13-amd64)
   AMD Radeon RX 7900 XTX (radeonsi, navi31, LLVM 19.1.7, DRM 3.59, 6.13.1-arch1-1)
*/
const QueryGameGPUTrend = `
SELECT
    report_month,
    gpu,
    COUNT(*) AS total_reports
FROM reports
WHERE steam_app_id = ?
GROUP BY report_month, gpu
ORDER BY report_month, total_reports DESC;
`

// same as above but for cpu instead of gpu and game specific too
const QueryGameCPUTrend = `
SELECT
    report_month,
    cpu,
    COUNT(*) AS total_reports
FROM reports
WHERE steam_app_id = ?
GROUP BY report_month, cpu
ORDER BY report_month, total_reports DESC;
`

// felt like useless because proton versions are changing and rarely did i see a game that required older proton versions to run
// game and date specific query for protonversions per month of a game
const QueryGameQuarterProtonVersions = `
SELECT
    report_month,
    proton_version,
    COUNT(*) AS total_reports
FROM reports
WHERE steam_app_id = ?
  AND report_date BETWEEN ? AND ?
GROUP BY report_month, proton_version
ORDER BY report_month, total_reports DESC;
`
