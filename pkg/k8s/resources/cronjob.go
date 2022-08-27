package resources

import (
	"kubectlfzf/pkg/util"
	"strings"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/batch/v1"
)

// CronJob is the summary of a kubernetes cronJob
type CronJob struct {
	ResourceMeta
	Schedule     string
	LastSchedule string
	Containers   []string
}

// NewCronJobFromRuntime builds a cronJob from informer result
func NewCronJobFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	c := &CronJob{}
	c.FromRuntime(obj, config)
	return c
}

// FromRuntime builds object from the informer's result
func (c *CronJob) FromRuntime(obj interface{}, config CtorConfig) {
	cronJob := obj.(*v1.CronJob)
	logrus.Tracef("Reading meta %#v", cronJob)
	c.FromObjectMeta(cronJob.ObjectMeta, config)
	c.Schedule = strings.ReplaceAll(cronJob.Spec.Schedule, " ", "_")
	c.LastSchedule = ""
	if cronJob.Status.LastScheduleTime != nil {
		c.LastSchedule = util.TimeToAge(cronJob.Status.LastScheduleTime.Time)
	}

	spec := cronJob.Spec.JobTemplate.Spec.Template.Spec
	containers := spec.Containers
	containers = append(containers, spec.InitContainers...)
	c.Containers = make([]string, len(containers))
	for k, v := range containers {
		c.Containers[k] = v.Name
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (c *CronJob) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (c *CronJob) ToStrings() []string {
	lst := []string{
		c.Namespace,
		c.Name,
		c.Schedule,
		c.LastSchedule,
		util.JoinSlicesOrNone(c.Containers, ","),
		c.resourceAge(),
		c.labelsString(),
	}
	return util.DumpLines(lst)
}
