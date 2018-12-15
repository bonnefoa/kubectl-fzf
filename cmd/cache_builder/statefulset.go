package main

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
)

// StatefulSet is the summary of a kubernetes statefulset
type StatefulSet struct {
	ResourceMeta
	currentReplicas int
	replicas        int
	selectors       []string
}

// FromRuntime builds object from the informer's result
func (s *StatefulSet) FromRuntime(obj interface{}) {
	statefulset := obj.(*appsv1.StatefulSet)
	s.FromObjectMeta(statefulset.ObjectMeta)
	s.currentReplicas = int(statefulset.Status.CurrentReplicas)
	s.replicas = int(statefulset.Status.Replicas)
	s.selectors = JoinStringMap(statefulset.Spec.Selector.MatchLabels, ExcludedLabels, "=")
}

// HasChanged returns true if the resource's dump needs to be updated
func (s *StatefulSet) HasChanged(k K8sResource) bool {
	oldSts := k.(*StatefulSet)
	return (s.currentReplicas != oldSts.currentReplicas ||
		s.replicas != oldSts.replicas ||
		StringSlicesEqual(s.selectors, oldSts.selectors) ||
		StringMapsEqual(s.labels, oldSts.labels))
}

// Header generates the csv header for the resource
func (s *StatefulSet) Header() string {
	return "Namespace Name Replicas Selector Age Labels\n"
}

// ToString serializes the object to strings
func (s *StatefulSet) ToString() string {
	selectorList := JoinSlicesOrNone(s.selectors, ",")
	line := strings.Join([]string{s.namespace,
		s.name,
		fmt.Sprintf("%d/%d", s.currentReplicas, s.replicas),
		selectorList,
		s.resourceAge(),
		s.labelsString(),
	}, " ")
	return fmt.Sprintf("%s\n", line)
}
