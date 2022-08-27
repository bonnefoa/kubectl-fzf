package util

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
)

func FormatCompletion(lines []string) string {
	logrus.Info("formating completion")
	b := new(strings.Builder)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', tabwriter.StripEscape)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	w.Flush()
	return b.String()
}
