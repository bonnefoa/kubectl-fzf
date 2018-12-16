package main

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const NodeHeader = "Name Roles InstanceType Zone InternalIp Age Labels\n"

// Node is the summary of a kubernetes node
type Node struct {
	ResourceMeta
	roles        []string
	instanceType string
	zone         string
	internalIP   string
}

// NewNodeFromRuntime builds a k8sresoutce from informer result
func NewNodeFromRuntime(obj interface{}) K8sResource {
	n := &Node{}
	n.FromRuntime(obj)
	return n
}

// FromRuntime builds object from the informer's result
func (n *Node) FromRuntime(obj interface{}) {
	node := obj.(*corev1.Node)
	n.FromObjectMeta(node.ObjectMeta)
	for k, v := range n.labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			n.roles = append(n.roles, v)
		}
	}
	n.instanceType = n.labels["beta.kubernetes.io/instance-type"]
	n.zone = n.labels["failure-domain.beta.kubernetes.io/zone"]
	for _, v := range node.Status.Addresses {
		if v.Type == "InternalIP" {
			n.internalIP = v.Address
		}
	}
}

// HasChanged returns true if the resource's dump needs to be updated
func (n *Node) HasChanged(k K8sResource) bool {
	return true
}

// ToString serializes the object to strings
func (n *Node) ToString() string {
	line := strings.Join([]string{n.name,
		JoinSlicesOrNone(n.roles, ","),
		n.instanceType,
		n.zone,
		n.internalIP,
		n.resourceAge(),
		n.labelsString(),
	}, " ")
	return fmt.Sprintf("%s\n", line)
}
