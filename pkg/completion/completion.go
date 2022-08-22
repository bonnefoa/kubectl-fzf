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
	cmdUse string, args []string) ([]string, []string, error) {
	var comps []string
	var err error
	resourceType := getResourceType(cmdUse, args)
	logrus.Debugf("Call Get Fun with %+v, resource type detected %s", args, resourceType)

	if resourceType == resources.ResourceTypeUnknown {
		return nil, comps, UnknownResourceError(strings.Join(args, " "))
	}

	labelHeader := []string{"Cluster", "Namespace", "Label", "Occurrences"}
	fieldSelectorHeader := []string{"Cluster", "Namespace", "FieldSelector", "Occurrences"}

	namespace := getNamespace(args)
	if len(args) >= 2 {
		lastWord := args[len(args)-1]
		penultimateWord := args[len(args)-2]
		logrus.Debugf("Checking lastWord '%s' and penultimateWord '%s'", lastWord, penultimateWord)
		if penultimateWord == "-l" || penultimateWord == "--selector" || lastWord == "-l" || lastWord == "-l=" || lastWord == "--selector=" || lastWord == "--selector" {
			comps, err := GetTagResourceCompletion(ctx, resourceType, namespace, fetchConfig, TagTypeLabel)
			return labelHeader, comps, err
		}
		if penultimateWord == "--field-selector" || lastWord == "--field-selector" || lastWord == "--field-selector=" {
			comps, err := GetTagResourceCompletion(ctx, resourceType, namespace, fetchConfig, TagTypeFieldSelector)
			return fieldSelectorHeader, comps, err
		}
	}

	header := resources.ResourceToHeader(resourceType)
	comps, err = getResourceCompletion(ctx, resourceType, namespace, fetchConfig)
	if err != nil {
		return header, comps, errors.Wrap(err, "error getting resource completion")
	}
	sort.Strings(comps)
	return header, comps, err
}

func ProcessCommandArgs(cmdUse string, args []string) (string, []string, error) {
	fetchConfigCli := fetcher.GetFetchConfigCli()
	f := fetcher.NewFetcher(&fetchConfigCli)
	err := f.SetClusterNameFromCurrentContext()
	if err != nil {
		return "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	header, comps, err := processCommandArgsWithFetchConfig(ctx, f, cmdUse, args)
	cancel()
	return util.DumpLine(header), comps, err
}
