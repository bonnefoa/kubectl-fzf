package resources

import (
	"fmt"
	"kubectlfzf/pkg/util"

	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
)

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
	logrus.Tracef("Reading meta %#v", job)
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
func (j *Job) ToStrings() []string {
	lst := []string{
		j.Namespace,
		j.Name,
		j.Completions,
		util.JoinSlicesOrNone(j.Containers, ","),
		j.resourceAge(),
		j.labelsString(),
	}
	return util.DumpLines(lst)
}
