package completion

import (
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"
)

func getApiCompletion(storeConfig *k8sresources.StoreConfig) []string {
	apiResources := []k8sresources.APIResource{}
	err := util.LoadFromFile(&apiResources, storeConfig.GetFilePath(k8sresources.ResourceTypeApiResource))
	util.FatalIf(err)
	res := []string{}
	res = append(res, k8sresources.APIResourceHeader)
	for _, v := range apiResources {
		res = append(res, v.ToString())
	}
	return res
}

// GetResourceCompletion gets the list of the resource specified which begin with `toComplete`.
func GetResourceCompletion(r k8sresources.ResourceType, storeConfig *k8sresources.StoreConfig) []string {
	if r == k8sresources.ResourceTypeApiResource {
		return getApiCompletion(storeConfig)
	}
	resources := map[string]k8sresources.K8sResource{}
	err := util.LoadFromFile(&resources, storeConfig.GetFilePath(r))
	util.FatalIf(err)
	res := []string{}
	res = append(res, k8sresources.ResourceToHeader(r))
	for _, v := range resources {
		res = append(res, v.ToString())
	}
	return res
}
