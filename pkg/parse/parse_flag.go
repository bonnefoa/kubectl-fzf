package parse

import (
	"strings"

	"github.com/sirupsen/logrus"
)

type FlagCompletion int

const (
	FlagLabel FlagCompletion = iota
	FlagFieldSelector
	FlagNamespace
	FlagNone
	FlagUnmanaged
)

func (f FlagCompletion) String() string {
	flagStr := [...]string{"Label", "FieldSelector", "Namespace", "None", "Unmanaged"}
	if len(flagStr) < int(f) {
		return "Unknown"
	}
	return flagStr[f]
}

func parsePreviousFlag(s string) FlagCompletion {
	logrus.Debugf("Parsing previous flag '%s'", s)
	switch s {
	case "l":
		return FlagLabel
	case "selector":
		return FlagLabel
	case "field-selector":
		return FlagFieldSelector
	case "n":
		fallthrough
	case "namespace":
		return FlagNamespace
	}
	return FlagNone
}

func parseLastFlag(s string) FlagCompletion {
	logrus.Debugf("Parsing last flag '%s'", s)
	switch s {
	case "l":
		fallthrough
	case "l=":
		fallthrough
	case "selector=":
		return FlagLabel
	case "n":
		fallthrough
	case "n=":
		fallthrough
	case "namespace=":
		return FlagNamespace
	case "field-selector=":
		return FlagFieldSelector
	}
	return FlagUnmanaged
}

func CheckFlagManaged(args []string) FlagCompletion {
	logrus.Infof("Checking Managed Flag '%s'", args)
	lastArg := args[len(args)-1]
	if strings.HasPrefix(lastArg, "-") {
		lastArg = strings.TrimLeft(lastArg, "-")
		return parseLastFlag(lastArg)
	}
	if len(args) >= 2 {
		penultimateArg := args[len(args)-2]
		if strings.HasPrefix(penultimateArg, "-") {
			penultimateArg = strings.TrimLeft(penultimateArg, "-")
			return parsePreviousFlag(penultimateArg)
		}
	}
	return FlagNone
}
