package util

import (
	"fmt"
	"regexp"
	"strings"
)

// IsStringExcluded returns true if one of the regexp match the input string
func IsStringExcluded(s string, regexps []*regexp.Regexp) bool {
	for _, regexp := range regexps {
		if regexp.MatchString(s) {
			return true
		}
	}
	return false
}

// IsStringIncluded returns true if one of the regexp match the input string
func IsStringIncluded(s string, regexps []*regexp.Regexp) bool {
	if len(regexps) == 0 {
		return true
	}
	for _, regexp := range regexps {
		if regexp.MatchString(s) {
			return true
		}
	}
	return false
}

// DumpLine replaces empty string by None, join the slice and append newline
func DumpLine(lst []string) []string {
	for k, v := range lst {
		if v == "" {
			lst[k] = "None"
		}
	}
	line := strings.Join(lst, " ")
	return []string{line}
}

// StringSlicesEqual returns true if slices are equals
func StringSlicesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a[k] != b[k] {
			return false
		}
	}
	return true
}

// StringMapsEqual returns true if maps are equals
func StringMapsEqual(a map[string]string, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a[k] != b[k] {
			return false
		}
	}
	return true
}

// JoinSlicesWithMaxOrNone joins a slice of string with separator up to x elements. Display None if there's no elements
func JoinSlicesWithMaxOrNone(sl []string, max int, sep string) string {
	if len(sl) == 0 {
		return "None"
	}
	if len(sl) < max {
		return strings.Join(sl, sep)
	}
	toDisplay := sl[:max]
	toDisplay = append(toDisplay, "...")
	return strings.Join(toDisplay, sep)
}

// JoinSlicesOrNone joins a slice of string with separator or display None if there's no elements
func JoinSlicesOrNone(sl []string, sep string) string {
	if len(sl) == 0 {
		return "None"
	}
	return strings.Join(sl, sep)
}

// TruncateString truncates a string to a given maximum
func TruncateString(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

// StringSliceToSet converts a string slice to a set like map
func StringSliceToSet(lst []string) map[string]bool {
	res := make(map[string]bool)
	for _, el := range lst {
		res[el] = true
	}
	return res
}

// JoinStringMap generates a list of map element separated by string excluding keys in excluded maps
func JoinStringMap(m map[string]string, exclude map[string]string, sep string) []string {
	res := make([]string, 0)
	for k, v := range m {
		if _, ok := exclude[k]; ok {
			continue
		}
		res = append(res, fmt.Sprintf("%s%s%s", k, sep, v))
	}
	return res
}

// LastURLPart extracts the last part of the url
func LastURLPart(url string) string {
	urlArray := strings.Split(url, "/")
	return urlArray[len(urlArray)-1]
}

func StringSliceToRegexps(s []string) ([]*regexp.Regexp, error) {
	res := make([]*regexp.Regexp, len(s))
	for i, ns := range s {
		rg, err := regexp.Compile(ns)
		if err != nil {
			return nil, err
		}
		res[i] = rg
	}
	return res, nil
}

func IsStringMatching(s string, regexps []*regexp.Regexp) bool {
	for _, regexp := range regexps {
		if regexp.MatchString(s) {
			return true
		}
	}
	return false
}
