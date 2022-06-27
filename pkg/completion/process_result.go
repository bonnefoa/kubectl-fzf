package completion

import (
	"strings"

	flag "github.com/spf13/pflag"
)

func ProcessResult(fzfResult string, initialCommand string) (string, error) {
	return "", nil
}

func processResultWithNamespace(fzfResult string, initialCommand string, currentNamespace string) (string, error) {
	strings.Split(fzfResult, " ")
	fs := flag.NewFlagSet("f1", flag.ContinueOnError)
	cmdNamespace := fs.StringP("namespace", "n", "", "")
	fs.Parse(strings.Split(initialCommand, " "))
	if cmdNamespace != nil && *cmdNamespace != currentNamespace {
		return fzfResult
	}

	//if namespace == currentNamespace {
	//}
	//for _, el := range strings.Split(initialCommand, " ") {
	//if strings.HasPrefix("--namespace=") || strings.HasPrefix("-n=")

	//}
	return "", nil
}
