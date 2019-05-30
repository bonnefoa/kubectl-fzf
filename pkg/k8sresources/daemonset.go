package k8sresources

import (
	"fmt"
	"strconv"

	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	"github.com/golang/glog"
	betav1 "k8s.io/api/extensions/v1beta1"
)

// DaemonSetHeader is the header file for daemonset
const DaemonSetHeader = "Namespace Name Desired Current Ready LabelSelector Containers Age Labels\n"

// DaemonSet is the summary of a kubernetes daemonset
type DaemonSet struct {
	ResourceMeta
	desired       string
	current       string
	ready         string
	containers    []string
	labelSelector []string
}

// NewDaemonSetFromRuntime builds a daemonset from informer result
func NewDaemonSetFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	d := &DaemonSet{}
	d.FromRuntime(obj, config)
	return d
}

// FromRuntime builds object from the informer's result
func (d *DaemonSet) FromRuntime(obj interface{}, config CtorConfig) {
	daemonset := obj.(*betav1.DaemonSet)
	glog.V(19).Infof("Reading meta %#v", daemonset)
	d.FromObjectMeta(daemonset.ObjectMeta)

	status := daemonset.Status
	d.desired = strconv.Itoa(int(status.DesiredNumberScheduled))
	d.current = strconv.Itoa(int(status.CurrentNumberScheduled))
	d.ready = strconv.Itoa(int(status.NumberReady))

	d.labelSelector = make([]string, 0)
	for k, v := range daemonset.Spec.Selector.MatchLabels {
		d.labelSelector = append(d.labelSelector, fmt.Sprintf("%s=%s", k, v))
	}

	podSpec := daemonset.Spec.Template.Spec
	containers := podSpec.Containers
	containers = append(containers, podSpec.InitContainers...)
	d.containers = make([]string, len(containers))
	for k, v := range containers {
		d.containers[k] = v.Name
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (d *DaemonSet) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (d *DaemonSet) ToString() string {
	lst := []string{
		d.namespace,
		d.name,
		d.desired,
		d.current,
		d.ready,
		util.JoinSlicesOrNone(d.labelSelector, ","),
		util.JoinSlicesOrNone(d.containers, ","),
		d.resourceAge(),
		d.labelsString(),
	}
	return util.DumpLine(lst)
}
