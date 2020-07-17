package resourcewatcher

import (
	"context"
	"io/ioutil"
	"kubectlfzf/pkg/util"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type K8sAggregator struct {
	storeConfig     StoreConfig
	ch              chan string
	resourceName    string
	aggregatedChans []chan string
	header          string
	destFileName    string
}

func NewK8sAggregator(cfg WatchConfig, storeConfig StoreConfig,
	aggregatedChans []chan string, ch chan string) (K8sAggregator, error) {

	k := K8sAggregator{}
	k.aggregatedChans = aggregatedChans
	destFileName := util.GetDestFileName(storeConfig.CacheDir, storeConfig.Cluster, cfg.resourceName)
	k.destFileName = destFileName
	k.resourceName = cfg.resourceName
	k.header = cfg.header
	k.storeConfig = storeConfig
	k.ch = ch

	util.WriteHeaderFile(cfg.header, destFileName)
	return k, nil
}

func (k *K8sAggregator) generateResourceOutput() (string, error) {
	var res strings.Builder
	for _, c := range k.aggregatedChans {
		c <- "resource"
	}
	for _, c := range k.aggregatedChans {
		output := <-c
		res.WriteString(output)
	}
	return res.String(), nil
}

func (k *K8sAggregator) writeAggregatedFile() error {
	str, err := k.generateResourceOutput()
	if err != nil {
		return errors.Wrapf(err, "Error when generating output for %s",
			k.resourceName)
	}

	tempFile, err := ioutil.TempFile(k.storeConfig.CacheDir, k.resourceName)
	if err != nil {
		return errors.Wrapf(err, "Error creating temp file for resource %s",
			k.resourceName)
	}

	err = util.WriteStringToFile(str, tempFile)
	if err != nil {
		return err
	}

	err = os.Rename(tempFile.Name(), k.destFileName)

	return nil
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
				output, err = k.generateResourceOutput()
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
			k.writeAggregatedFile()
		}
	}
}
