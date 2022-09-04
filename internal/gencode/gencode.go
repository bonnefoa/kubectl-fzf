package gencode

import (
	"context"
	"fmt"
	"strings"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
)

func outputResourceType(apiResource resources.APIResource) {
	parsedResource := resources.ParseResourceType(apiResource.Name)
	if parsedResource == resources.ResourceTypeUnknown {
		return
	}

	for _, shortName := range apiResource.Shortnames {
		fmt.Printf("\tcase \"%s\":\n", shortName)
		fmt.Print("\t\tfallthrough\n")
	}
	fmt.Printf("\tcase \"%s\":\n", strings.ToLower(apiResource.Kind))
	fmt.Print("\t\tfallthrough\n")
	fmt.Printf("\tcase \"%s\":\n", apiResource.Name)
	fmt.Printf("\t\treturn ResourceType%s\n", apiResource.Kind)
}

func outputNamespaced(apiResource resources.APIResource) {
	parsedResource := resources.ParseResourceType(apiResource.Name)
	if parsedResource == resources.ResourceTypeUnknown {
		return
	}
	fmt.Printf("\tcase ResourceType%s:\n", apiResource.Kind)
	fmt.Printf("\t\treturn %t\n", apiResource.Namespaced)
}

func GenerateResourceCode(ctx context.Context) error {
	clusterConfigCli := clusterconfig.GetClusterConfigCli()
	clusterConfig := clusterconfig.NewClusterConfig(clusterConfigCli)
	err := clusterConfig.LoadClusterConfig()
	if err != nil {
		return err
	}
	clientSet, err := clusterConfig.GetClientset()
	if err != nil {
		return err
	}
	resourceLists, err := clientSet.Discovery().ServerPreferredResources()
	if err != nil {
		return err
	}

	fmt.Print("Api resource:\n")
	for _, resourceList := range resourceLists {
		apiResourceList := resources.APIResourceList{}
		apiResourceList.FromRuntime(resourceList, resources.CtorConfig{})
		for _, apiResource := range apiResourceList.ApiResources {
			outputResourceType(apiResource)
		}
	}

	fmt.Print("Namespaced:\n")
	for _, resourceList := range resourceLists {
		apiResourceList := resources.APIResourceList{}
		apiResourceList.FromRuntime(resourceList, resources.CtorConfig{})
		for _, apiResource := range apiResourceList.ApiResources {
			outputNamespaced(apiResource)
		}
	}

	return nil
}
