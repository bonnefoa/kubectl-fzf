package completion

import (
	"context"
	"kubectlfzf/pkg/fetcher"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/parse"
	"kubectlfzf/pkg/util"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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
	var comps []string
	var err error
	resourceType, flagCompletion, err := parse.ParseFlagAndResources(cmdVerb, args)
	if err != nil {
		return nil, nil, err
	}
	logrus.Debugf("Call Get Fun with %+v, resource type detected %s, flag detected %s", args, resourceType, flagCompletion)
	labelHeader := []string{"Cluster", "Namespace", "Label", "Occurrences"}
	fieldSelectorHeader := []string{"Cluster", "Namespace", "FieldSelector", "Occurrences"}

	namespace := getNamespace(args)
	if flagCompletion == parse.FlagLabel {
		comps, err := GetTagResourceCompletion(ctx, resourceType, namespace, fetchConfig, TagTypeLabel)
		return labelHeader, comps, err
	} else if flagCompletion == parse.FlagFieldSelector {
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
