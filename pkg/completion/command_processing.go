package completion

import (
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func getNamespace(args []string) *string {
	for k, arg := range args {
		if arg == "-n" || arg == "--namespace" {
			return &args[k+1]
		}
		if strings.HasPrefix(arg, "--namespace=") {
			return &strings.Split(arg, "=")[1]
		}
	}
	return nil
}

func ProcessCommandArgs(args []string) ([]string, error) {
	clusterCliConf := util.GetClusterCliConf()
	storeConfig := k8sresources.NewStoreConfig(&clusterCliConf, 0)

	var comps []string
	var err error
	logrus.Debugf("Call Get Fun with %+v", args)
	resourceType := k8sresources.ResourceTypeApiResource
	// No resource type or we have only
	// k get #
	if len(args) < 2 {
		return GetResourceCompletion(k8sresources.ResourceTypeApiResource, nil, storeConfig)
	}

	for _, v := range args {
		resourceType = k8sresources.ParseResourceType(v)
		if resourceType != k8sresources.ResourceTypeUnknown {
			break
		}
	}

	if resourceType == k8sresources.ResourceTypeUnknown {
		logrus.Debugf("Couldn't find resource type in '%v'", args)
		os.Exit(1)
	}
	namespace := getNamespace(args)
	comps, err = GetResourceCompletion(resourceType, namespace, storeConfig)
	return comps, err
}

func ProcessResults(initialCommand string, results []string) (string, error) {
	return "", nil
}
