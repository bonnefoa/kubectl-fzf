package parse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmanagedArgs(t *testing.T) {
	cmdArgs := [][]string{
		{"-t"},
		{"-i"},
		{"--field-selector"},
		{"--selector"},
	}
	for _, args := range cmdArgs {
		r := CheckFlagManaged(args)
		require.Equal(t, FlagUnmanaged.String(), r.String())
	}
}

type flagTest struct {
	flag   []string
	result FlagCompletion
}

func TestManagedArgs(t *testing.T) {
	cmdArgs := []flagTest{
		{[]string{"--selector="}, FlagLabel},
		{[]string{"--field-selector", ""}, FlagFieldSelector},
		{[]string{"--field-selector="}, FlagFieldSelector},
		{[]string{"--all-namespaces", ""}, FlagNone},
		{[]string{"-t", ""}, FlagNone},
		{[]string{"-i", ""}, FlagNone},
		{[]string{"-ti", ""}, FlagNone},
		{[]string{"-it", ""}, FlagNone},
		{[]string{"-n"}, FlagNamespace},
		{[]string{"-n="}, FlagNamespace},
		{[]string{"-n", " "}, FlagNamespace},
		{[]string{"--namespace", ""}, FlagNamespace},
	}
	for _, args := range cmdArgs {
		r := CheckFlagManaged(args.flag)
		require.Equal(t, args.result.String(), r.String())
	}
}
