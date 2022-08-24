package completion

import (
	"context"
	"kubectlfzf/pkg/fetcher"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/util"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type UnknownResourceError string

func (u UnknownResourceError) Error() string {
	return string(u)
}

type UnmanagedFlagError string

func (u UnmanagedFlagError) Error() string {
	return string(u)
}

func getNamespace(args []string) *string {
	for k, arg := range args {
		if (arg == "-n" || arg == "--namespace") && len(args) > k+1 {
			return &args[k+1]
		}
		if strings.HasPrefix(arg, "--namespace=") {
			return &strings.Split(arg, "=")[1]
		}
	}
	return nil
}

func getResourceCompletion(ctx context.Context, r resources.ResourceType, namespace *string,
	fetchConfig *fetcher.Fetcher) ([]string, error) {
	resources, err := fetchConfig.GetResources(ctx, r)
	if err != nil {
		return nil, err
	}
	comps := []string{}
	logrus.Debugf("Filterting with namespace %v", namespace)
	for _, resource := range resources {
		if namespace == nil || *namespace == resource.GetNamespace() {
			comps = append(comps, resource.ToStrings()...)
		}
	}
	return comps, nil
}

func processCommandArgsWithFetchConfig(ctx context.Context, fetchConfig *fetcher.Fetcher,
	cmdVerb string, args []string) ([]string, []string, error) {
	flagCompletion := checkFlagManaged(args)
	if flagCompletion == FlagUnmanaged {
		logrus.Infof("Flag is unmanaged in %s, bailing out", args)
		return nil, nil, UnmanagedFlagError(strings.Join(args, " "))
	}
	var comps []string
	var err error
	resourceType := getResourceType(cmdVerb, args)
	logrus.Debugf("Call Get Fun with %+v, resource type detected %s, flag detected %s", args, resourceType, flagCompletion)

	if resourceType == resources.ResourceTypeUnknown {
		return nil, comps, UnknownResourceError(strings.Join(args, " "))
	}

	labelHeader := []string{"Cluster", "Namespace", "Label", "Occurrences"}
	fieldSelectorHeader := []string{"Cluster", "Namespace", "FieldSelector", "Occurrences"}

	namespace := getNamespace(args)
	if flagCompletion == FlagLabel {
		comps, err := GetTagResourceCompletion(ctx, resourceType, namespace, fetchConfig, TagTypeLabel)
		return labelHeader, comps, err
	} else if flagCompletion == FlagFieldSelector {
		comps, err := GetTagResourceCompletion(ctx, resourceType, namespace, fetchConfig, TagTypeFieldSelector)
		return fieldSelectorHeader, comps, err
	}

	header := resources.ResourceToHeader(resourceType)
	comps, err = getResourceCompletion(ctx, resourceType, namespace, fetchConfig)
	if err != nil {
		return header, comps, errors.Wrap(err, "error getting resource completion")
	}
	sort.Strings(comps)
	return header, comps, err
}

func ProcessCommandArgs(cmdVerb string, args []string) (string, []string, error) {
	fetchConfigCli := fetcher.GetFetchConfigCli()
	f := fetcher.NewFetcher(&fetchConfigCli)
	err := f.SetClusterNameFromCurrentContext()
	if err != nil {
		return "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	header, comps, err := processCommandArgsWithFetchConfig(ctx, f, cmdVerb, args)
	cancel()
	return util.DumpLine(header), comps, err
}
