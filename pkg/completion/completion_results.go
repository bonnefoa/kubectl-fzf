package completion

import (
	"fmt"
	"kubectlfzf/pkg/util"
)

type CompletionResult struct {
	Cluster     string
	Header      string
	Completions []string
}

func (c *CompletionResult) GetFormattedOutput() string {
	lines := []string{fmt.Sprintf("Cluster: %s", c.Cluster), c.Header}
	lines = append(lines, c.Completions...)
	return util.FormatCompletion(lines)
}
