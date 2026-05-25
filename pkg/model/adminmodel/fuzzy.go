package adminmodel

import "regexp"

func FuzzyQueryIsId(fuzzyQuery string) bool {
	if fuzzyQuery == "" {
		return false
	}
	return regexp.MustCompile(`^-?\d+(\.\d+)?$`).MatchString(fuzzyQuery)
}
