package models

type Game struct {
	ID                  string
	Title               string
	SteamDeckStatus     string
	VerdictDistribution map[string]int
	CalculatedRating    string
	OfficialRating      string
	TotalReports        int
}

/*
To check
Verdict = Yes and VerdictOutOfTheBox = No => Not Recommended Vanilla, Recommend to Tinker
Verdict = Yes and VerdictOutOfTheBox = Yes => Recommended to play but need to Tinker
Verdict = Yes => Recommended Vanilla, No need to Tinker
Verdict = No => Not Working, Not Recommended
*/
type Report struct {
	Installs            bool
	ProtonVersion       string
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
