package k8sresources

import (
	"kubectlfzf/pkg/util"
	"os"
	"path"
	"time"
)

// StoreConfig defines parameters used for the cache location
type StoreConfig struct {
	cacheDir            string
	destDir             string
	TimeBetweenFullDump time.Duration
}

func NewStoreConfig(clusterConf *util.ClusterCliConf, timeBetweenFullDump time.Duration) *StoreConfig {
	s := StoreConfig{}
	clusterName := clusterConf.GetClusterName()
	s.cacheDir = clusterConf.CacheDir
	s.destDir = path.Join(s.cacheDir, clusterName)
	s.TimeBetweenFullDump = timeBetweenFullDump

	err := os.MkdirAll(s.destDir, os.ModePerm)
	util.FatalIf(err)

	return &s
}

func (s *StoreConfig) GetFilePath(resourceName string) string {
	return path.Join(s.destDir, resourceName)
}
