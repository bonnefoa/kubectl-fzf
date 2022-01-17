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

// CompGetResource gets the list of the resource specified which begin with `toComplete`.
func CompGetResource(cmd *cobra.Command, resourceName string, toComplete string) []string {
	glog.V(3).Infof("Receive Get resource of resource %s with %s", resourceName, toComplete)
	//if resourceName == "pod" {
	//pods := []k8sresources.Pod{}
	//f := storeConfig.GetFilePath(resourceName)
	//err := util.LoadFromFile(pods, f)
	//util.FatalIf(err)
	//}
	return []string{}
}

func CompGetApiResources(cmd *cobra.Command, storeConfig *k8sresources.StoreConfig) []string {
	apiResources := []k8sresources.APIResource{}
	err := util.LoadFromFile(&apiResources, storeConfig.GetFilePath("apiresources"))
	util.FatalIf(err)
	res := []string{}
	for _, v := range apiResources {
		res = append(res, v.ToString())
	}
	return res
}

func completeGetFun(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	clusterCliConf := util.GetClusterCliConf()
	storeConfig := k8sresources.NewStoreConfig(&clusterCliConf, 0)

	var comps []string
	glog.V(3).Infof("Call Get Fun with %s and %s", args, toComplete)
	if len(args) == 0 {
		comps = CompGetApiResources(cmd, storeConfig)
		return comps, cobra.ShellCompDirectiveNoFileComp
	}
	//comps = CompGetResource(cmd, args[0], toComplete)
	//comps = cmdutil.Difference(comps, args[1:])
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func main() {
	var rootCmd = &cobra.Command{
		Use:               "completion __complete get pods",
		Short:             "Completion helper for kubectl",
		ValidArgsFunction: completeGetFun,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
