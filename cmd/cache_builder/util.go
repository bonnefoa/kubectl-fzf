package main

import (
	"fmt"
	"strconv"
	"strings"
)

// JoinStringMap generates a list of map element separated by string excluding keys in excluded maps
func JoinStringMap(m map[string]string, exclude map[string]string, sep string) []string {
	res := make([]string, len(m))
	i := 0
	for k, v := range m {
		res[i] = fmt.Sprintf("%s%s%s", k, sep, v)
		i++
	}
	return res
}

// JoinSlicesOrNone joins a slice of string with separator or display None if there's no elements
func JoinSlicesOrNone(sl []string, sep string) string {
	if len(sl) == 0 {
		return "None"
	}
	return strings.Join(sl, sep)
}

// ExcludeFromSlice removes elements in exclude map from slice sl
func ExcludeFromSlice(sl []string, exclude map[string]string) []string {
	res := make([]string, len(sl))
	i := 0
	for k, v := range sl {
		_, isExcluded := ExcludedLabels[v]
		if isExcluded {
			continue
		}
		res[k] = v
		i++
	}
	return res[:i]
}

// JoinIntSlice creates a string of joined int with a separator character
func JoinIntSlice(a []int, sep string) string {
	if len(a) == 0 {
		return "None"
	}
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}
	return strings.Join(b, sep)
}
