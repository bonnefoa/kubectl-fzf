package completion

import (
	"kubectlfzf/pkg/k8s/resources"

	"github.com/sirupsen/logrus"
)

func getResourceType(cmdUse string, args []string) resources.ResourceType {
	logrus.Debugf("Getting resource type from '%s', %d", args, len(args))
	resourceType := resources.ResourceTypeApiResource
	if cmdUse == "logs" {
		return resources.ResourceTypePod
	}
	if cmdUse == "exec" {
		return resources.ResourceTypePod
	}
	// No resource type or we have only
	// get ''#
	if len(args) <= 1 {
		return resourceType
	}
	for _, v := range args {
		resourceType = resources.ParseResourceType(v)
		if resourceType != resources.ResourceTypeUnknown {
			return resourceType
		}
	}
	return resources.ResourceTypeUnknown
}
