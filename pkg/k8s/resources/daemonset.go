package resources

import (
	"fmt"
	"kubectlfzf/pkg/util"
	"strconv"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
)

// DaemonSet is the summary of a kubernetes daemonset
type DaemonSet struct {
	ResourceMeta
	Desired       string
	Current       string
	Ready         string
	Containers    []string
	LabelSelector []string
}

// NewDaemonSetFromRuntime builds a daemonset from informer result
func NewDaemonSetFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	d := &DaemonSet{}
	d.FromRuntime(obj, config)
	return d
}

// FromRuntime builds object from the informer's result
func (d *DaemonSet) FromRuntime(obj interface{}, config CtorConfig) {
	daemonset := obj.(*appsv1.DaemonSet)
	logrus.Tracef("Reading meta %#v", daemonset)
	d.FromObjectMeta(daemonset.ObjectMeta, config)

	status := daemonset.Status
	d.Desired = strconv.Itoa(int(status.DesiredNumberScheduled))
	d.Current = strconv.Itoa(int(status.CurrentNumberScheduled))
	d.Ready = strconv.Itoa(int(status.NumberReady))

	d.LabelSelector = make([]string, 0)
	for k, v := range daemonset.Spec.Selector.MatchLabels {
		d.LabelSelector = append(d.LabelSelector, fmt.Sprintf("%s=%s", k, v))
	}

	podSpec := daemonset.Spec.Template.Spec
	containers := podSpec.Containers
	containers = append(containers, podSpec.InitContainers...)
	d.Containers = make([]string, len(containers))
	for k, v := range containers {
		d.Containers[k] = v.Name
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (d *DaemonSet) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (d *DaemonSet) ToStrings() []string {
	lst := []string{
		d.Cluster,
		d.Namespace,
		d.Name,
		d.Desired,
		d.Current,
		d.Ready,
		util.JoinSlicesOrNone(d.LabelSelector, ","),
		util.JoinSlicesOrNone(d.Containers, ","),
		d.resourceAge(),
		d.labelsString(),
	}
	return util.DumpLines(lst)
}
