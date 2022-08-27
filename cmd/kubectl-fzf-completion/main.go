package main

import (
	"context"
	"fmt"
	"kubectlfzf/pkg/completion"
	"kubectlfzf/pkg/fetcher"
	"kubectlfzf/pkg/fzf"
	"kubectlfzf/pkg/k8s/store"
	"kubectlfzf/pkg/parse"
	"kubectlfzf/pkg/results"
	"kubectlfzf/pkg/util"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func completeFun(cmd *cobra.Command, args []string) {
	header, comps, err := completion.ProcessCommandArgs(cmd.Use, args)
	if e, ok := err.(parse.UnknownResourceError); ok {
		logrus.Warnf("Unknown resource type: %s", e)
		os.Exit(6)
	}
	if e, ok := err.(parse.UnmanagedFlagError); ok {
		logrus.Warnf("Unmanaged flag: %s", e)
		os.Exit(6)
	}
	if err != nil {
		logrus.Fatalf("Completion error: %s", err)
	}
	if len(comps) == 0 {
		logrus.Warn("No completion found")
		os.Exit(5)
	}
	formattedComps := util.FormatCompletion(header, comps)

	fzfResult, err := fzf.CallFzf(formattedComps, "")
	if err != nil {
		logrus.Fatalf("Call fzf error: %s", err)
	}
	res, err := results.ProcessResult(cmd.Use, args, fzfResult)
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
		Use:     "k8s_completion",
		Short:   "Subcommand grouping completion for kubectl cli verbs",
		Example: "kubectl-fzf-completion k8s_completion get pods \"\"",
	}
	rootCmd.AddCommand(k8sCmd)
	verbs := []string{"get", "exec", "logs", "label", "describe", "delete", "annotate", "edit"}
	for _, verb := range verbs {
		cmd := &cobra.Command{
			Use:                verb,
			Run:                completeFun,
			DisableFlagParsing: true,
			FParseErrWhitelist: cobra.FParseErrWhitelist{
				UnknownFlags: true,
			},
		}
		k8sCmd.AddCommand(cmd)
	}
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

func main() {
	var rootCmd = &cobra.Command{
		Use: "kubectl_fzf_completion",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	rootFlags := rootCmd.PersistentFlags()
	util.SetCommonCliFlags(rootFlags)
	err := viper.BindPFlags(rootFlags)
	util.FatalIf(err)

	addK8sCmd(rootCmd)
	addStatsCmd(rootCmd)

	util.ConfigureViper()
	cobra.OnInitialize(util.ConfigureLog)
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Root command failed: %v", err)
	}
}
