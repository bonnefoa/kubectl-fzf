package main

import (
	"fmt"
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func init() {
	util.SetClusterConfFlags()
	util.ParseFlags()
}

func fallbackComp(cmd *cobra.Command, args []string) {
	os.Exit(1)
}

func completeGetFun(cmd *cobra.Command, args []string) {
	clusterCliConf := util.GetClusterCliConf()
	storeConfig := k8sresources.NewStoreConfig(&clusterCliConf, 0)

	var comps []string
	glog.V(3).Infof("Call Get Fun with %s", args)
	if len(args) == 0 {
		comps = CompGetApiResources(cmd, storeConfig)
		return comps
	}
	comps = CompGetResource(cmd, args[0], storeConfig)
	return comps
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "kubectl_completion",
		Run: fallbackComp,
	}

	verbs := []string{"get", "label", "describe", "delete", "annotate", "edit"}
	for _, verb := range verbs {
		var cmd = &cobra.Command{
			Use: verb,
			Run: completeGetFun,
		}
		rootCmd.AddCommand(cmd)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
