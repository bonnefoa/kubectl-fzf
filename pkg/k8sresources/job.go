package k8sresources

import (
	"strconv"

	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	"github.com/golang/glog"
	batchv1 "k8s.io/api/batch/v1"
)

// JobHeader is the headers for job files
const JobHeader = "Namespace Name Schedule LastSchedule Containers Age Labels\n"

// Job is the summary of a kubernetes Job
type Job struct {
	ResourceMeta
	desired    string
	successful string
	containers []string
}

// NewJobFromRuntime builds a Job from informer result
func NewJobFromRuntime(obj interface{}) K8sResource {
	j := &Job{}
	j.FromRuntime(obj)
	return j
}

// FromRuntime builds object from the informer's result
func (j *Job) FromRuntime(obj interface{}) {
	job := obj.(*batchv1.Job)
	glog.V(19).Infof("Reading meta %#v", job)
	j.FromObjectMeta(job.ObjectMeta)
	j.desired = strconv.Itoa(int(*job.Spec.Completions))
	j.successful = strconv.Itoa(int(job.Status.Succeeded))

	spec := job.Spec.Template.Spec
	containers := spec.Containers
	containers = append(containers, spec.InitContainers...)
	j.containers = make([]string, len(containers))
	for k, v := range containers {
		j.containers[k] = v.Name
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (j *Job) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (j *Job) ToString() string {
	lst := []string{
		j.namespace,
		j.name,
		j.desired,
		j.successful,
		util.JoinSlicesOrNone(j.containers, ","),
		j.resourceAge(),
		j.labelsString(),
	}
	return util.DumpLine(lst)
}
