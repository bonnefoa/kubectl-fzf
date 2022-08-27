package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"kubectlfzf/pkg/daemon"
	"kubectlfzf/pkg/httpserver"
	"kubectlfzf/pkg/k8s/resourcewatcher"
	"kubectlfzf/pkg/k8s/store"
	"kubectlfzf/pkg/kubectlfzfserver"
	"kubectlfzf/pkg/util"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	Version   string
	BuildDate string
	GitCommit string
	GitBranch string
	GoVersion string
)

func versionFun(cmd *cobra.Command, args []string) {
	fmt.Printf("Version: %s\n", Version)
	if GitCommit != "" {
		fmt.Printf("Git hash: %s\n", GitCommit)
	}
	if GitBranch != "" {
		fmt.Printf("Git branch: %s\n", GitBranch)
	}
	if BuildDate != "" {
		fmt.Printf("Build date: %s\n", BuildDate)
	}
	if GoVersion != "" {
		fmt.Printf("Go Version: %s\n", GoVersion)
	}
	os.Exit(0)
}

func startDaemonFun(cmd *cobra.Command, args []string) {
	daemon.StartDaemon()
}

func kubectlFzfServerFun(cmd *cobra.Command, args []string) {
	kubectlfzfserver.StartKubectlFzfServer()
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "kubectl_fzf_server",
		Run: kubectlFzfServerFun,
	}
	rootFlags := rootCmd.PersistentFlags()
	store.SetStoreConfigCli(rootFlags)
	httpserver.SetHttpServerConfigFlags(rootFlags)
	resourcewatcher.SetResourceWatcherCli(rootFlags)
	util.SetCommonCliFlags(rootFlags, "info")
	err := viper.BindPFlags(rootFlags)
	util.FatalIf(err)

	versionCmd := &cobra.Command{
		Use:   "version",
		Run:   versionFun,
		Short: "Print command version",
	}
	rootCmd.AddCommand(versionCmd)

	daemonCmd := &cobra.Command{
		Use: "daemon",
		Run: startDaemonFun,
	}
	daemonFlags := daemonCmd.Flags()
	daemon.SetDaemonFlags(daemonFlags)
	rootCmd.AddCommand(daemonCmd)
	err = viper.BindPFlags(daemonFlags)
	util.FatalIf(err)

	util.ConfigureViper()
	cobra.OnInitialize(util.ConfigureLog)
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Root command failed: %v", err)
	}
}
