package resourcewatcher

import (
	"context"
	"kubectlfzf/pkg/util"
	"path"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type K8sAggregator struct {
	storeConfig   StoreConfig
	ch            chan string
	resourceChans []chan string
	labelChans    []chan LabelPairList
	header        string
	destDir       string
	resourceName  string
}

func NewK8sAggregator(cfg WatchConfig, storeConfig StoreConfig,
	resourceChans []chan string, ch chan string) (K8sAggregator, error) {

	k := K8sAggregator{}
	k.resourceChans = resourceChans
	k.destDir = path.Join(storeConfig.CacheDir, storeConfig.Cluster)
	k.resourceName = cfg.resourceName
	k.header = cfg.header
	k.storeConfig = storeConfig
	k.ch = ch

	err := util.WriteStringToFile(cfg.header, k.destDir, k.resourceName, "header")
	return k, err
}

func (k *K8sAggregator) generateOutput(resourceType string) (string, error) {
	var res strings.Builder
	for _, c := range k.resourceChans {
		c <- resourceType
	}
	for _, c := range k.resourceChans {
		output := <-c
		res.WriteString(output)
	}
	return res.String(), nil
}

func (k *K8sAggregator) writeAggregatedFile(resourceType string) error {
	str, err := k.generateOutput(resourceType)
	if err != nil {
		return errors.Wrapf(err, "Error when generating output for %s",
			k.resourceName)
	}

	err = util.WriteStringToFile(str, k.destDir, k.resourceName, resourceType)
	return err
}

func (k *K8sAggregator) Start(ctx context.Context) {
	timer := time.NewTimer(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			glog.Infof("Exiting aggregator %s", k.resourceName)
			return

		case query := <-k.ch:
			var output string
			var err error
			if query == "resource" {
				output, err = k.generateOutput("resource")
			} else if query == "label" {
				output, err = k.generateOutput("label")
			} else if query == "header" {
				output = k.header
			} else {
				k.ch <- "Invalid query"
			}

			if err != nil {
				k.ch <- "Error"
			} else {
				k.ch <- output
			}

		case <-timer.C:
			k.writeAggregatedFile("resource")
			k.writeAggregatedFile("label")
		}
	}
}
