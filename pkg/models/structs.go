package models

type Game struct {
	ID                  string
	Title               string
	SteamDeckStatus     string
	VerdictDistribution map[string]int
	CalculatedRating    string
	OfficialRating      string
	TotalReports        int
	TopProtonVersions   []VersionStats
	TopLinuxDistros     []DistroStats
}

type VersionStats struct {
	Version    string
	Count      int
	Percentage float64
}

type DistroStats struct {
	Distro     string
	Count      int
	Percentage float64
}

/*
To check
Verdict = Yes and VerdictOutOfTheBox = No => Not Recommended Vanilla, Recommend to Tinker
Verdict = Yes and VerdictOutOfTheBox = Yes => Recommended to play but need to Tinker
Verdict = Yes => Recommended Vanilla, No need to Tinker
Verdict = No => Not Working, Not Recommended
should we update ProtonVer and CustomProton names?
*/
type Report struct {
	AppTitle      string
	Installs      bool
	ProtonVersion string
	// todo
	CustomProtonVersion string
	Verdict             string
	VerdictOutOfTheBox  string
	OperatingSystem     string
	ImpactedByAntiCheat bool
}

type EventFlag struct {
	GameRatingChange    string
	CommonProtonVersion string
	LinuxDistroSpread   string
	NoRatingChanges     string
}
