package completion

import (
	"errors"
	"fmt"
	"kubectlfzf/pkg/k8s/fetcher"
	"kubectlfzf/pkg/k8s/resources"
	"sort"

	"golang.org/x/net/context"
)

type TagType int64

const (
	TagTypeLabel TagType = iota
	TagTypeFieldSelector
)

type TagResourceKey struct {
	Cluster   string
	Namespace string
	Value     string
}

type TagResourcePair struct {
	Key         TagResourceKey
	Occurrences int
}

type TagResourcePairList []TagResourcePair

func (p TagResourcePairList) Len() int { return len(p) }
func (p TagResourcePairList) Less(i, j int) bool {
	if p[i].Occurrences == p[j].Occurrences {
		if p[i].Key.Cluster == p[j].Key.Cluster {
			if p[i].Key.Namespace == p[j].Key.Namespace {
				return p[i].Key.Value < p[j].Key.Value
			}
			return p[i].Key.Namespace < p[j].Key.Namespace
		}
		return p[i].Key.Cluster < p[j].Key.Cluster
	}
	return p[i].Occurrences > p[j].Occurrences
}
func (p TagResourcePairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (l *TagResourcePair) ToString() string {
	return fmt.Sprintf("%s\t%s\t%s\t%d", l.Key.Cluster, l.Key.Namespace,
		l.Key.Value, l.Occurrences)
}

func getTagResourceOccurrences(ctx context.Context, r resources.ResourceType, namespace *string,
	fetchConfig *fetcher.Fetcher, tagType TagType) (map[TagResourceKey]int, error) {
	if r == resources.ResourceTypeApiResource {
		return nil, errors.New("no map resource completion on api resource")
	}
	resources, err := fetchConfig.GetResources(ctx, r)
	if err != nil {
		return nil, err
	}
	resourceKeyToOccurrences := make(map[TagResourceKey]int, 0)
	for _, resource := range resources {
		if namespace == nil || *namespace == resource.GetNamespace() {
			var tagResource map[string]string
			if tagType == TagTypeLabel {
				tagResource = resource.GetLabels()
			} else {
				tagResource = resource.GetFieldSelectors()
			}
			for k, v := range tagResource {
				valueStr := fmt.Sprintf("%s=%s", k, v)
				valueKey := TagResourceKey{resource.GetCluster(), resource.GetNamespace(), valueStr}
				resourceKeyToOccurrences[valueKey] += 1
			}
		}
	}
	return resourceKeyToOccurrences, nil
}

func GetTagResourceCompletion(ctx context.Context, r resources.ResourceType, namespace *string,
	fetchConfig *fetcher.Fetcher, tagType TagType) ([]string, error) {
	tagResourceOccurrencesMap, err := getTagResourceOccurrences(ctx, r, namespace, fetchConfig, tagType)
	if err != nil {
		return nil, err
	}
	tagResourcePairList := make(TagResourcePairList, 0)
	for k, occurrence := range tagResourceOccurrencesMap {
		tagResourcePairList = append(tagResourcePairList, TagResourcePair{k, occurrence})
	}
	sort.Sort(tagResourcePairList)
	labelComps := make([]string, 0)
	for _, labelPair := range tagResourcePairList {
		labelComps = append(labelComps, labelPair.ToString())
	}
	return labelComps, nil
}
