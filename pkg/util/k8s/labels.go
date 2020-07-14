package k8s

import (
	"log"
	"regexp"
)

var labelAntiSelector *regexp.Regexp

func init() {
	var err error

	// Regex matching all non-allowed label characters
	labelAntiSelector, err = regexp.Compile("[^a-z0-9A-Z_.-]+")
	if err != nil {
		log.Fatal(err)
	}
}

// Make given string label-compatible
func StringToLabel(s string) string {
	// We need to remove characters that are not allowed
	s = labelAntiSelector.ReplaceAllString(s, "_")

	// We need to trim the string to allowed length
	if len(s) > 63 {
		s = s[0:63]
	}

	return s
}
