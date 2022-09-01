package store

import (
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type StoreConfigCli struct {
	*clusterconfig.ClusterConfigCli
	TimeBetweenFullDump time.Duration
}

func SetStoreConfigCli(fs *pflag.FlagSet) {
	clusterconfig.SetClusterConfigCli(fs)
	fs.Duration("time-between-full-dump", 10*time.Second, "Buffer changes and only do full dump every x secondes")
}

func GetStoreConfigCli() StoreConfigCli {
	s := StoreConfigCli{
		ClusterConfigCli: clusterconfig.GetClusterConfigCli(),
	}
	s.TimeBetweenFullDump = viper.GetDuration("time-between-full-dump")
	return s
}
