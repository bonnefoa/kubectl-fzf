package k8sresources

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APIResourceHeader is the header for pod files
const APIResourceHeader = "Fullname Shortnames GroupVersion Namespaced Kind\n"

// APIResource is the summary of a kubernetes pod
type APIResource struct {
	FullName     string
	Shortnames   []string
	GroupVersion string
	Namespaced   bool
	Kind         string
}

// FromAPIResource builds object from discovery query
func (a *APIResource) FromAPIResource(apiResource metav1.APIResource,
	resourceList *metav1.APIResourceList) {
	a.Shortnames = apiResource.ShortNames
	a.GroupVersion = resourceList.GroupVersion
	a.Namespaced = apiResource.Namespaced
	a.Kind = apiResource.Kind

	a.FullName = apiResource.Name
	if strings.Contains(a.GroupVersion, "/") {
		group := strings.Split(a.GroupVersion, "/")[0]
		a.FullName = fmt.Sprintf("%s.%s", apiResource.Name, group)
	}
}
