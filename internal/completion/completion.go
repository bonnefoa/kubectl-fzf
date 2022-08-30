package completion

import (
	"context"
	"sort"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/parse"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const labelHeader = "Namespace\tLabel\tOccurrences"
const fieldSelectorHeader = "Namespace\tFieldSelector\tOccurrences"

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
	cmdVerb string, args []string) (*CompletionResult, error) {
	var err error
	resourceType, flagCompletion, err := parse.ParseFlagAndResources(cmdVerb, args)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Call Get Fun with %+v, resource type detected %s, flag detected %s", args, resourceType, flagCompletion)

	completionResult := &CompletionResult{Cluster: fetchConfig.GetContext()}
	namespace := parse.ParseNamespaceFromArgs(args)
	if flagCompletion == parse.FlagLabel {
		completionResult.Completions, err = GetTagResourceCompletion(ctx, resourceType, namespace, fetchConfig, TagTypeLabel)
		completionResult.Header = labelHeader
		return completionResult, err
	} else if flagCompletion == parse.FlagFieldSelector {
		completionResult.Completions, err = GetTagResourceCompletion(ctx, resourceType, namespace, fetchConfig, TagTypeFieldSelector)
		completionResult.Header = fieldSelectorHeader
		return completionResult, err
	}

	completionResult.Header = resources.ResourceToHeader(resourceType)
	completionResult.Completions, err = getResourceCompletion(ctx, resourceType, namespace, fetchConfig)
	if err != nil {
		return completionResult, errors.Wrap(err, "error getting resource completion")
	}
	sort.Strings(completionResult.Completions)
	return completionResult, err
}

func ProcessCommandArgs(cmdVerb string, args []string, f *fetcher.Fetcher) (*CompletionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	completionResult, err := processCommandArgsWithFetchConfig(ctx, f, cmdVerb, args)
	cancel()
	return completionResult, err
}
