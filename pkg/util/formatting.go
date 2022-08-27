package util

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
)

func FormatCompletion(header string, comps []string) string {
	logrus.Info("formating completion")
	b := new(strings.Builder)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', tabwriter.StripEscape)
	fmt.Fprintln(w, header)
	for _, c := range comps {
		fmt.Fprintln(w, c)
	}
	w.Flush()
	return b.String()
}
