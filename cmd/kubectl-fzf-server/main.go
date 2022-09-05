package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/httpserver"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resourcewatcher"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/kubectlfzfserver"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	version   = "dev"
	gitCommit = "none"
	gitBranch = "unknown"
	goVersion = "unknown"
	buildDate = "unknown"
)

func versionFun(cmd *cobra.Command, args []string) {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Git hash: %s\n", gitCommit)
	fmt.Printf("Git branch: %s\n", gitBranch)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Go Version: %s\n", goVersion)
	os.Exit(0)
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

	util.ConfigureViper()
	cobra.OnInitialize(util.CommonInitialization)
	defer pprof.StopCPUProfile()
	defer util.DoMemoryProfile()
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Root command failed: %v", err)
	}
}
