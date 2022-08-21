package store

import (
	"fmt"
	"kubectlfzf/pkg/k8s/resources"
	"strings"
	"text/tabwriter"
	"time"
)

type Stats struct {
	ResourceType     resources.ResourceType
	ItemPerNamespace map[string]int
	LastDumped       time.Time
}

func GetStatsFromStores(stores []*Store) []*Stats {
	stats := make([]*Stats, 0)
	for _, stores := range stores {
		stats = append(stats, stores.GetStats())
	}
	return stats
}

func (s *Stats) toTabOutput() []string {
	strings := make([]string, 0)
	now := time.Now()
	for namespace, numItems := range s.ItemPerNamespace {
		deltaDate := now.Sub(s.LastDumped).Truncate(time.Second)
		line := fmt.Sprintf("%s\t%s\t%d\t%s",
			s.ResourceType.String(),
			namespace,
			numItems,
			deltaDate,
		)
		strings = append(strings, line)
	}
	return strings
}

func GetStatsOutput(stats []*Stats) string {
	b := new(strings.Builder)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', tabwriter.StripEscape)
	fmt.Fprintln(w, "Resource\tNamespace\tNumber\tLast Dumped")
	for _, s := range stats {
		for _, line := range s.toTabOutput() {
			fmt.Fprintln(w, line)
		}
	}
	w.Flush()
	return b.String()
}
