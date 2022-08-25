package completion

import (
	"fmt"
	"kubectlfzf/pkg/fetcher"
	"kubectlfzf/pkg/k8s/resources"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// ProcessResult handles fzf output and provides completion to use
// The fzfResult should have the first 3 columns of the fzf preview
func ProcessResult(cmdUse string, cmdArgs []string, fzfResult string) (string, error) {
	logrus.Debugf("Processing fzf result %s", fzfResult)
	logrus.Debugf("Cmd command %s", cmdArgs)
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
	return processResultWithNamespace(cmdUse, cmdArgs, fzfResult, namespace)
}

func parseNamespaceFlag(cmdArgs []string) (*string, error) {
	fs := pflag.NewFlagSet("f1", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	cmdNamespace := fs.StringP("namespace", "n", "", "")
	logrus.Debugf("Parsing namespace from %v", cmdArgs)
	err := fs.Parse(cmdArgs)
	return cmdNamespace, err
}

func processResultWithNamespace(cmdUse string, cmdArgs []string, fzfResult string, currentNamespace string) (string, error) {
	// If apiresource:
	// 0 -> fullname, 1 -> shortname, 2 -> groupversion
	// If namespace:
	// 0 -> cluster, 1 -> name, 2 -> age
	// Otherwise:
	// 0 -> cluster, 1 -> namespace, 2 -> value
	resultFields := strings.Fields(fzfResult)
	if len(resultFields) < 3 {
		return "", fmt.Errorf("fzf result should have at least 3 elements, got %v", resultFields)
	}
	logrus.Debugf("Processing fzfResult '%s', cmdArgs '%s', current namespace '%s'", fzfResult, cmdArgs, currentNamespace)
	resourceType := getResourceType(cmdUse, cmdArgs)
	if resourceType == resources.ResourceTypeUnknown {
		return "", fmt.Errorf("unkonwn resource type from source command %s", cmdArgs)
	}
	logrus.Debugf("Resource type found: %s", resourceType)
	if resourceType == resources.ResourceTypeApiResource {
		return resultFields[0], nil
	}
	if resourceType == resources.ResourceTypeNamespace {
		return resultFields[1], nil
	}

	// Generic resource
	resultNamespace := resultFields[1]
	resultValue := resultFields[2]
	logrus.Debugf("Result namespace: %s, resultValue: %s", resultNamespace, resultValue)

	cmdNamespace, err := parseNamespaceFlag(cmdArgs)
	if err != nil {
		return "", errors.Wrapf(err, "Error parsing commands %s", cmdArgs)
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
