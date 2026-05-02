package scraper

// on hold
func DetectRatingChange(currentRating, previousRating string) bool {
	if currentRating != previousRating {
		return true
	}
	return false

}
