package main

import (
	"fmt"
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"
	"os"

	"github.com/spf13/cobra"
)

// CompGetResource gets the list of the resource specified which begin with `toComplete`.
func CompGetResource(cmd *cobra.Command, resourceName string, toComplete string) []string {
	storeConfig := k8sresources.NewStoreConfig()
	if resourceName == "pod" {
		pod := k8sresources.Pod{}
		util.LoadFromFile(p)
	}
	return []string{}
}

func completeGetFun(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var comps []string
	if len(args) == 0 {
		return comps, cobra.ShellCompDirectiveNoFileComp
	}
	comps = CompGetResource(cmd, args[0], toComplete)
	comps = cmdutil.Difference(comps, args[1:])
}

func main() {
	var rootCmd = &cobra.Command{
		Use:               "kubectl-fzf",
		Short:             "Completion helper for kubectl",
		ValidArgsFunction: completeGetFun,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
