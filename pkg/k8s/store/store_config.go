package store

import (
	"kubectlfzf/pkg/k8s/clusterconfig"
	"time"
)

// StoreConfig defines configuration to store
// This is shared between all resources
type StoreConfig struct {
	clusterconfig.ClusterConfig
	timeBetweenFullDump time.Duration
}

func NewStoreConfig(storeConfigCli *StoreConfigCli) *StoreConfig {
	s := StoreConfig{}
	s.ClusterConfig = clusterconfig.NewClusterConfig(&storeConfigCli.ClusterConfigCli)
	s.timeBetweenFullDump = storeConfigCli.TimeBetweenFullDump
	return &s
}

func (s *StoreConfig) GetTimeBetweenFullDump() time.Duration {
	return s.timeBetweenFullDump
}
