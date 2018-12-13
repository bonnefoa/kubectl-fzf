package main

// Pod is the summary of a kubernetes pod
type Pod struct {
	ResourceMeta
	hostIP     string
	podIP      string
	nodeName   string
	containers []string
	phase      string
}
