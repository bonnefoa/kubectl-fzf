package k8sresources

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APIResourceHeader is the header for pod files
const APIResourceHeader = "Fullname Shortnames GroupVersion Namespaced Kind\n"

// APIResource is the summary of a kubernetes pod
type APIResource struct {
	fullName     string
	shortnames   []string
	groupVersion string
	namespaced   bool
	kind         string
}

// FromAPIResource builds object from discovery query
func (a *APIResource) FromAPIResource(apiResource metav1.APIResource,
	resourceList *metav1.APIResourceList) {
	a.shortnames = apiResource.ShortNames
	a.groupVersion = resourceList.GroupVersion
	a.namespaced = apiResource.Namespaced
	a.kind = apiResource.Kind

	a.fullName = apiResource.Name
	if strings.Contains(a.groupVersion, "/") {
		group := strings.Split(a.groupVersion, "/")[0]
		a.fullName = fmt.Sprintf("%s.%s", apiResource.Name, group)
	}
}

// ToString serializes the object to strings
func (a *APIResource) ToString() string {
	lst := []string{
		a.fullName,
		util.JoinSlicesOrNone(a.shortnames, ","),
		a.groupVersion,
		strconv.FormatBool(a.namespaced),
		a.kind,
	}
	return util.DumpLine(lst)
}
