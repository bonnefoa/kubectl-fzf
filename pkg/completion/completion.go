package completion

import (
	"kubectlfzf/pkg/k8sresources"
	"kubectlfzf/pkg/util"

	"github.com/sirupsen/logrus"
)

func getApiCompletion(storeConfig *k8sresources.StoreConfig) ([]string, error) {
	apiResources := []k8sresources.APIResource{}
	err := util.LoadFromFile(&apiResources, storeConfig.GetFilePath(k8sresources.ResourceTypeApiResource))
	if err != nil {
		return nil, err
	}
	res := []string{}
	res = append(res, k8sresources.APIResourceHeader)
	for _, v := range apiResources {
		res = append(res, v.ToString())
	}
	return res, nil
}

// GetResourceCompletion gets the list of the resource specified which begin with `toComplete`.
func GetResourceCompletion(r k8sresources.ResourceType, namespace *string,
	storeConfig *k8sresources.StoreConfig) ([]string, error) {
	if r == k8sresources.ResourceTypeApiResource {
		return getApiCompletion(storeConfig)
	}
	resources := map[string]k8sresources.K8sResource{}
	err := util.LoadFromFile(&resources, storeConfig.GetFilePath(r))
	if err != nil {
		return nil, err
	}
	res := []string{}
	res = append(res, k8sresources.ResourceToHeader(r))
	logrus.Debugf("Filterting with namespace %v", namespace)
	for _, resource := range resources {
		if namespace == nil || *namespace == resource.GetNamespace() {
			res = append(res, resource.ToString())
		}
	}
	return res, nil
}
