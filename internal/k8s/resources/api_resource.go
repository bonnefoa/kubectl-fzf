package resources

import (
	"strconv"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APIResource is the summary of a kubernetes pod
type APIResourceList struct {
	ApiResources []APIResource
	GroupVersion string
}

// FromRuntime builds object from the informer's result
func (r *APIResourceList) FromRuntime(obj interface{}, config CtorConfig) {
	resourceList := obj.(*metav1.APIResourceList)
	r.GroupVersion = resourceList.GroupVersion
	for _, apiResource := range resourceList.APIResources {
		a := APIResource{}
		a.Shortnames = apiResource.ShortNames
		a.Namespaced = apiResource.Namespaced
		a.Kind = apiResource.Kind
		a.Name = apiResource.Name
		a.Version = r.GroupVersion
		r.ApiResources = append(r.ApiResources, a)
	}
}

func (r *APIResourceList) HasChanged(k K8sResource) bool {
	return true
}

func (r *APIResourceList) GetNamespace() string {
	return ""
}

func (r *APIResourceList) GetLabels() map[string]string {
	return nil
}

func (r *APIResourceList) GetFieldSelectors() map[string]string {
	return nil
}

// ToString serializes the object to strings
func (r *APIResourceList) ToStrings() []string {
	lines := []string{}
	for _, a := range r.ApiResources {
		lines = append(lines, a.ToStrings()...)
	}
	return lines
}

// APIResource is the summary of a kubernetes pod
type APIResource struct {
	Name       string
	Shortnames []string
	Version    string
	Namespaced bool
	Kind       string
}

func (a *APIResource) ToStrings() []string {
	lst := []string{
		a.Name,
		util.JoinSlicesOrNone(a.Shortnames, ","),
		a.Version,
		strconv.FormatBool(a.Namespaced),
		a.Kind,
	}
	return util.DumpLines(lst)
}
