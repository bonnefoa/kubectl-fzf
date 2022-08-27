package resources

import (
	"fmt"
	"sort"
	"strings"

	"kubectlfzf/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

// Node is the summary of a kubernetes node
type Node struct {
	ResourceMeta
	Roles        []string
	Status       string
	InstanceType string
	Zone         string
	InstanceID   string
	InternalIP   string
	Taints       []string
}

// NewNodeFromRuntime builds a k8sresoutce from informer result
func NewNodeFromRuntime(obj interface{}, config CtorConfig) K8sResource {
	n := &Node{}
	n.FromRuntime(obj, config)
	return n
}

func getNodeStatus(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status != "True" {
				return condition.Reason
			}
		}
	}
	return "Ready"
}

// FromRuntime builds object from the informer's result
func (n *Node) FromRuntime(obj interface{}, config CtorConfig) {
	node := obj.(*corev1.Node)
	n.FromObjectMeta(node.ObjectMeta, config)
	for k := range n.Labels {
		nodePrefix := "node-role.kubernetes.io/"
		if strings.HasPrefix(k, nodePrefix) {
			role := strings.Replace(k, nodePrefix, "", 1)
			if _, ok := config.IgnoredNodeRoles[role]; ok {
				continue
			}
			n.Roles = append(n.Roles, role)
		}
	}
	n.InstanceID = "Unknown"
	if node.Spec.ProviderID != "" {
		n.InstanceID = util.LastURLPart(node.Spec.ProviderID)
	}

	n.Status = getNodeStatus(node)

	n.Taints = make([]string, 0)
	for _, t := range node.Spec.Taints {
		var taint string
		if t.Value == "" {
			taint = fmt.Sprintf("%s:%s", t.Key, t.Effect)
		} else {
			taint = fmt.Sprintf("%s=%s:%s", t.Key, t.Value, t.Effect)
		}
		n.Taints = append(n.Taints, taint)
	}

	n.InstanceType = n.Labels["beta.kubernetes.io/instance-type"]
	n.Zone = n.Labels["failure-domain.beta.kubernetes.io/zone"]
	for _, v := range node.Status.Addresses {
		if v.Type == "InternalIP" {
			n.InternalIP = v.Address
		}
	}
	sort.Strings(n.Roles)
}

// HasChanged returns true if the resource's dump needs to be updated
func (n *Node) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (n *Node) ToStrings() []string {
	line := []string{
		n.Name,
		util.JoinSlicesOrNone(n.Roles, ","),
		n.Status,
		n.InstanceType,
		n.Zone,
		n.InternalIP,
		util.JoinSlicesOrNone(n.Taints, ","),
		n.InstanceID,
		n.resourceAge(),
		n.labelsString(),
	}
	return util.DumpLines(line)
}
