package k8sresources

import (
	"fmt"
	"kubectlfzf/pkg/util"
	"os"
	"path"
	"time"
)

// StoreConfig defines parameters used for the cache location
type StoreConfig struct {
	clusterDir          string
	cacheDir            string
	destDir             string
	TimeBetweenFullDump time.Duration
}

func NewStoreConfig(clusterConf *util.ClusterCliConf, timeBetweenFullDump time.Duration) *StoreConfig {
	s := StoreConfig{}
	s.clusterDir = clusterConf.GetClusterDir()
	s.cacheDir = clusterConf.CacheDir
	s.destDir = path.Join(s.cacheDir, s.clusterDir)
	s.TimeBetweenFullDump = timeBetweenFullDump

	err := os.MkdirAll(s.destDir, os.ModePerm)
	util.FatalIf(err)

	return &s
}

func (s *StoreConfig) GetFilePath(suffix string, resourceName string) string {
	return path.Join(s.destDir, fmt.Sprintf("%s_%s", resourceName, "resource"))
}
