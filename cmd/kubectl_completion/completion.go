package main

import (
	"fmt"
	"kubectlfzf/pkg/completion"
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	util.SetClusterConfFlags()
	util.SetLogConfFlags()

	util.ParseFlags()

	util.ConfigureLog()
}

func fallbackComp(cmd *cobra.Command, args []string) {
	logrus.Debugf("Fallback comp %s", args)
	os.Exit(1)
}

func completeGetFun(cmd *cobra.Command, args []string) {
	clusterCliConf := util.GetClusterCliConf()
	storeConfig := k8sresources.NewStoreConfig(&clusterCliConf, 0)

	var comps []string
	logrus.Debugf("Call Get Fun with %+v", args)
	resourceType := k8sresources.ResourceTypeApiResource
	// No resource type or we have only
	// k get #
	// k get pod#
	if len(args) < 2 {
		comps = completion.CompGetApiResources(storeConfig)
	} else {
		resourceType = k8sresources.ParseResourceType(args[0])
		if resourceType == k8sresources.ResourceTypeUnknown {
			logrus.Debugf("Resource Type '%v' unknown", resourceType, args[0])
			os.Exit(1)
		}
		comps = completion.CompGetResource(resourceType, storeConfig)
	}
	fmt.Print(strings.Join(comps, ""))
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
		logrus.Fatalf("Root command failed", err)
	}
}
