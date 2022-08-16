package completion

import (
	"fmt"
	"kubectlfzf/pkg/k8s/fetcher"
	"kubectlfzf/pkg/k8s/resources"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// ProcessResult handles fzf output and provides completion to use
// The fzfResult should have the first 3 columns of the fzf preview
func ProcessResult(fzfResult string, sourceCommand string) (string, error) {
	logrus.Debugf("Processing fzf result %s", fzfResult)
	logrus.Debugf("Source command %s", sourceCommand)
	fetchConfigCli := fetcher.GetFetchConfigCli()
	fetcher := fetcher.NewFetcher(&fetchConfigCli)
	namespace := ""
	err := fetcher.SetClusterNameFromCurrentContext()
	if err != nil {
		logrus.Debugf("Error building fetcher: %v, falling back to empty namespace", err)
	} else {
		namespace, err = fetcher.GetNamespace()
		if err != nil {
			logrus.Debugf("Error getting namespace: %v, falling back to empty namespace", err)
		}
	}
	return processResultWithNamespace(fzfResult, sourceCommand, namespace)
}

func parseNamespaceFlag(sourceCommand string) (*string, error) {
	fs := pflag.NewFlagSet("f1", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	cmdNamespace := fs.StringP("namespace", "n", "", "")
	splitCommand := strings.Split(sourceCommand, " ")
	logrus.Debugf("Parsing namespace from %v", splitCommand)
	err := fs.Parse(splitCommand)
	return cmdNamespace, err
}

func processResultWithNamespace(fzfResult string, sourceCommand string, currentNamespace string) (string, error) {
	// If apiresource:
	// 0 -> fullname, 1 -> shortname, 2 -> groupversion
	// If namespace:
	// 0 -> cluster, 1 -> name, 2 -> age
	// Otherwise:
	// 0 -> cluster, 1 -> namespace, 2 -> value
	fzfResultSplits := strings.Split(fzfResult, " ")
	if len(fzfResultSplits) != 3 {
		return "", fmt.Errorf("fzf result should have 3 elements, got %v", fzfResultSplits)
	}
	cmdArgs := strings.Split(sourceCommand, " ")
	logrus.Debugf("Processing fzfResult '%s', sourceCommand '%s', current namespace '%s'", fzfResult, sourceCommand, currentNamespace)
	resourceType := getResourceType(cmdArgs[0], cmdArgs[1:])
	if resourceType == resources.ResourceTypeUnknown {
		return "", fmt.Errorf("unkonwn resource type from source command %s", sourceCommand)
	}
	if resourceType == resources.ResourceTypeApiResource {
		return fzfResultSplits[0], nil
	}
	if resourceType == resources.ResourceTypeNamespace {
		return fzfResultSplits[1], nil
	}

	// Generic resource
	resultNamespace := fzfResultSplits[1]
	resultValue := fzfResultSplits[2]

	cmdNamespace, err := parseNamespaceFlag(sourceCommand)
	if err != nil {
		return "", errors.Wrapf(err, "Error parsing commands %s", sourceCommand)
	}
	lastWord := cmdArgs[len(cmdArgs)-1]
	// add flag to the completion
	if lastWord == "-l=" || lastWord == "-l" || lastWord == "--field-selector=" || lastWord == "--selector=" {
		resultValue = fmt.Sprintf("%s%s", lastWord, resultValue)
	}

	if cmdNamespace != nil && *cmdNamespace == resultNamespace {
		return resultValue, nil
	}

	if resultNamespace != currentNamespace {
		completion := fmt.Sprintf("%s -n %s", resultValue, resultNamespace)
		return completion, nil
	}
	return resultValue, nil
}
