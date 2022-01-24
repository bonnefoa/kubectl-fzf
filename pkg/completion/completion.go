package completion

import (
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"
)

func CompGetApiResources(storeConfig *k8sresources.StoreConfig) []string {
	apiResources := []k8sresources.APIResource{}
	err := util.LoadFromFile(&apiResources, storeConfig.GetFilePath("apiresources"))
	util.FatalIf(err)
	res := []string{}
	res = append(res, k8sresources.APIResourceHeader)
	for _, v := range apiResources {
		res = append(res, v.ToString())
	}
	return res
}

// CompGetResource gets the list of the resource specified which begin with `toComplete`.
func CompGetResource(resourceName string, storeConfig *k8sresources.StoreConfig) []string {
	resources := map[string]k8sresources.K8sResource{}
	err := util.LoadFromFile(&resources, storeConfig.GetFilePath(resourceName))
	util.FatalIf(err)
	res := []string{}
	for _, v := range resources {
		res = append(res, v.ToString())
	}
	return res
}