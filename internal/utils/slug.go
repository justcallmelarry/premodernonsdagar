package utils

import (
	"regexp"
	"strings"
)

func Slugify(s string) string {
	s = strings.ToLower(s)

	reg := regexp.MustCompile("[^a-z0-9 ]+")
	s = reg.ReplaceAllString(s, "")

	s = strings.ReplaceAll(s, " ", "-")

	reg = regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")

	return s
}
