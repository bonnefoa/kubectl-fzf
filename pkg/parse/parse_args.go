package parse

import (
	"kubectlfzf/pkg/k8s/resources"
	"strings"

	"github.com/sirupsen/logrus"
)

type UnmanagedFlagError string

func (u UnmanagedFlagError) Error() string {
	return string(u)
}

type UnknownResourceError string

func (u UnknownResourceError) Error() string {
	return string(u)
}

func ParseFlagAndResources(cmdVerb string, cmdArgs []string) (resourceType resources.ResourceType, flagCompletion FlagCompletion, err error) {
	resourceType = resources.ResourceTypeUnknown
	flagCompletion = CheckFlagManaged(cmdArgs)
	if flagCompletion == FlagUnmanaged {
		logrus.Infof("Flag is unmanaged in %s, bailing out", cmdArgs)
		err = UnmanagedFlagError(strings.Join(cmdArgs, " "))
		return
	}

	if flagCompletion == FlagNamespace {
		resourceType = resources.ResourceTypeNamespace
		return
	}
	resourceType = resources.GetResourceType(cmdVerb, cmdArgs)

	if resourceType == resources.ResourceTypeUnknown {
		err = UnknownResourceError(strings.Join(cmdArgs, " "))
		return
	}
	return
}
