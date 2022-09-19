package main

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/completion"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/fzf"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/gencode"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/parse"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/results"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const FallbackExitCode = 6

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

func completeFun(cmd *cobra.Command, cmdArgs []string) {
	args := completion.PrepareCmdArgs(cmdArgs)
	if args == nil {
		os.Exit(FallbackExitCode)
	}

	firstWord := args[0]
	verbs := []string{"get", "exec", "logs", "label", "describe", "delete", "annotate", "edit", "scale"}
	if !util.IsStringIn(firstWord, verbs) {
		os.Exit(FallbackExitCode)
	}
	args = args[1:]

	fetchConfigCli := fetcher.GetFetchConfigCli()
	f := fetcher.NewFetcher(&fetchConfigCli)
	err := f.LoadFetcherState()
	if err != nil {
		logrus.Warnf("Error loading fetcher state")
		os.Exit(FallbackExitCode)
	}

	completionResults, err := completion.ProcessCommandArgs(firstWord, args, f)
	if e, ok := err.(resources.UnknownResourceError); ok {
		logrus.Warnf("Unknown resource type: %s", e)
		os.Exit(FallbackExitCode)
	} else if e, ok := err.(parse.UnmanagedFlagError); ok {
		logrus.Warnf("Unmanaged flag: %s", e)
		os.Exit(FallbackExitCode)
	} else if err != nil {
		logrus.Warnf("Error during completion: %s", err)
		os.Exit(FallbackExitCode)
	}

	err = f.SaveFetcherState()
	if err != nil {
		logrus.Warnf("Error saving fetcher state: %s", err)
		os.Exit(FallbackExitCode)
	}

	if err != nil {
		logrus.Fatalf("Completion error: %s", err)
	}
	if len(completionResults.Completions) == 0 {
		logrus.Warn("No completion found")
		os.Exit(5)
	}
	formattedComps := completionResults.GetFormattedOutput()

	// TODO pass query
	fzfResult, err := fzf.CallFzf(formattedComps, "")
	if err != nil {
		if e, ok := err.(fzf.InterruptedCommandError); ok {
			logrus.Infof("Fzf was interrupted: %s", e)
			os.Exit(FallbackExitCode)
		}
		logrus.Fatalf("Call fzf error: %s", err)
	}
	res, err := results.ProcessResult(firstWord, args, f, fzfResult)
	if err != nil {
		logrus.Fatalf("Process result error: %s", err)
	}
	fmt.Print(res)
}

func statsFun(cmd *cobra.Command, args []string) {
	fetchConfigCli := fetcher.GetFetchConfigCli()
	f := fetcher.NewFetcher(&fetchConfigCli)
	ctx := context.Background()
	stats, err := f.GetStats(ctx)
	util.FatalIf(err)
	statsOutput := store.GetStatsOutput(stats)
	fmt.Print(statsOutput)
}

func addK8sCmd(rootCmd *cobra.Command) {
	var k8sCmd = &cobra.Command{
		Use:                "k8s_completion",
		Run:                completeFun,
		Short:              "Subcommand grouping completion for kubectl cli verbs",
		Example:            "kubectl-fzf-completion k8s_completion get pods \"\"",
		DisableFlagParsing: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	rootCmd.AddCommand(k8sCmd)
}

func addStatsCmd(rootCmd *cobra.Command) {
	statsCmd := &cobra.Command{
		Use: "stats",
		Run: statsFun,
	}
	statsFlags := statsCmd.Flags()
	fetcher.SetFetchConfigFlags(statsFlags)
	err := viper.BindPFlags(statsFlags)
	util.FatalIf(err)
	rootCmd.AddCommand(statsCmd)
}

func genFun(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	err := gencode.GenerateResourceCode(ctx)
	util.FatalIf(err)
}

func addGenCommand(rootCmd *cobra.Command) {
	genCmd := &cobra.Command{
		Use: "generate",
		Run: genFun,
	}
	fs := genCmd.PersistentFlags()
	clusterconfig.SetClusterConfigCli(fs)
	rootCmd.AddCommand(genCmd)
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "kubectl_fzf_completion",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	rootFlags := rootCmd.PersistentFlags()
	util.SetCommonCliFlags(rootFlags, "error")
	err := viper.BindPFlags(rootFlags)
	util.FatalIf(err)

	versionCmd := &cobra.Command{
		Use:   "version",
		Run:   versionFun,
		Short: "Print command version",
	}
	rootCmd.AddCommand(versionCmd)

	addK8sCmd(rootCmd)
	addStatsCmd(rootCmd)
	addGenCommand(rootCmd)

	util.ConfigureViper()
	cobra.OnInitialize(util.CommonInitialization)
	defer pprof.StopCPUProfile()
	if err := rootCmd.Execute(); err != nil {
		logrus.Errorf("Root command failed: %v", err)
	}
}
