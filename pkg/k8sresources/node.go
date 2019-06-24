package k8sresources

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bonnefoa/kubectl-fzf/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

// NodeHeader is the header line of csv result
const NodeHeader = "Name Roles Status InstanceType Zone InternalIp Taints InstanceID Age Labels\n"

// Node is the summary of a kubernetes node
type Node struct {
	ResourceMeta
	roles        []string
	status       string
	instanceType string
	zone         string
	instanceID   string
	internalIP   string
	taints       []string
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
	n.FromObjectMeta(node.ObjectMeta)
	for k := range n.labels {
		nodePrefix := "node-role.kubernetes.io/"
		if strings.HasPrefix(k, nodePrefix) {
			role := strings.Replace(k, nodePrefix, "", 1)
			if _, ok := config.RoleBlacklist[role]; ok {
				continue
			}
			n.roles = append(n.roles, role)
		}
	}
	n.instanceID = "Unknown"
	if node.Spec.ProviderID != "" {
		n.instanceID = util.LastURLPart(node.Spec.ProviderID)
	}

	n.status = getNodeStatus(node)

	n.taints = make([]string, 0)
	for _, t := range node.Spec.Taints {
		var taint string
		if t.Value == "" {
			taint = fmt.Sprintf("%s:%s", t.Key, t.Effect)
		} else {
			taint = fmt.Sprintf("%s=%s:%s", t.Key, t.Value, t.Effect)
		}
		n.taints = append(n.taints, taint)
	}

	n.instanceType = n.labels["beta.kubernetes.io/instance-type"]
	n.zone = n.labels["failure-domain.beta.kubernetes.io/zone"]
	for _, v := range node.Status.Addresses {
		if v.Type == "InternalIP" {
			n.internalIP = v.Address
		}
	}
	sort.Strings(n.roles)
}

// HasChanged returns true if the resource's dump needs to be updated
func (n *Node) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (n *Node) ToString() string {
	line := strings.Join([]string{n.name,
		util.JoinSlicesOrNone(n.roles, ","),
		n.status,
		n.instanceType,
		n.zone,
		n.internalIP,
		util.JoinSlicesOrNone(n.taints, ","),
		n.instanceID,
		n.resourceAge(),
		n.labelsString(),
	}, " ")
	return fmt.Sprintf("%s\n", line)
}
