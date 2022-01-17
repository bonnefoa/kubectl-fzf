package k8sresources

import (
	"fmt"
	"kubectlfzf/pkg/util"

	"github.com/golang/glog"
	batchv1 "k8s.io/api/batch/v1"
)

// JobHeader is the headers for job files
const JobHeader = "Cluster Namespace Name Completions Containers Age Labels\n"

// Job is the summary of a kubernetes Job
type Job struct {
	ResourceMeta
	Completions string
	Containers  []string
}

// NewJobFromRuntime builds a Job from informer result
func NewJobFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	j := &Job{}
	j.FromRuntime(obj, config)
	return j
}

// FromRuntime builds object from the informer's result
func (j *Job) FromRuntime(obj interface{}, config CtorConfig) {
	job := obj.(*batchv1.Job)
	glog.V(19).Infof("Reading meta %#v", job)
	j.FromObjectMeta(job.ObjectMeta, config)

	j.Completions = "-"
	if job.Spec.Completions != nil {
		desired := int(*job.Spec.Completions)
		successful := int(job.Status.Succeeded)
		j.Completions = fmt.Sprintf("%d/%d", successful, desired)
	}

	spec := job.Spec.Template.Spec
	containers := spec.Containers
	containers = append(containers, spec.InitContainers...)
	j.Containers = make([]string, len(containers))
	for k, v := range containers {
		j.Containers[k] = v.Name
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (j *Job) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (j *Job) ToString() string {
	lst := []string{
		j.Cluster,
		j.Namespace,
		j.Name,
		j.Completions,
		util.JoinSlicesOrNone(j.Containers, ","),
		j.resourceAge(),
		j.labelsString(),
	}
	return util.DumpLine(lst)
}
