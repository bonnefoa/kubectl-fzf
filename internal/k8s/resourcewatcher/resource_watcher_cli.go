package resourcewatcher

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ResourceWatcherCli struct {
	watchedResources       []string
	excludedResources      []string
	watchedNamespaces      []string
	excludedNamespaces     []string
	ignoredNodeRoles       []string
	nodePollingPeriod      time.Duration
	namespacePollingPeriod time.Duration
}

func SetResourceWatcherCli(fs *pflag.FlagSet) {
	fs.StringSlice("watch-resources", []string{}, "Resources to watch, separated by comma.")
	fs.StringSlice("exclude-resources", []string{}, "Resources to exclude, separated by comma. To exclude everything: pods,configmaps,services,serviceaccounts,replicasets,daemonsets,secrets,statefulsets,deployments,endpoints,ingresses,cronjobs,jobs,horizontalpodautoscalers,persistentvolumes,persistentvolumeclaims,nodes,namespaces.")
	fs.StringSlice("watch-namespaces", []string{}, "Namespace regexps to watch, separated by comma.")
	fs.StringSlice("exclude-namespaces", []string{}, "Namespace regexps to exclude, separated by comma.")
	fs.StringSlice("ignore-node-roles", []string{}, "List of node role to ommit in the dump. It won't appaear in the completion. Useful to save space and remove cluster for 'common' node role. Separated by comma.")
	fs.Duration("node-polling-period", 300*time.Second, "Polling period for nodes.")
	fs.Duration("namespace-polling-period", 600*time.Second, "Polling period for namespaces.")
}

func GetResourceWatcherCli() ResourceWatcherCli {
	r := ResourceWatcherCli{}
	r.watchedResources = viper.GetStringSlice("watch-resources")
	r.watchedNamespaces = viper.GetStringSlice("watch-namespaces")
	r.excludedResources = viper.GetStringSlice("exclude-resources")
	r.excludedNamespaces = viper.GetStringSlice("exclude-namespaces")
	r.ignoredNodeRoles = viper.GetStringSlice("ignore-node-roles")
	r.nodePollingPeriod = viper.GetDuration("node-polling-period")
	r.namespacePollingPeriod = viper.GetDuration("namespace-polling-period")
	return r
}
