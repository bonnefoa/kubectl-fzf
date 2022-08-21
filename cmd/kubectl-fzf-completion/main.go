package main

import (
	"fmt"
	"kubectlfzf/pkg/completion"
	"kubectlfzf/pkg/k8s/fetcher"
	"kubectlfzf/pkg/util"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func completeFun(cmd *cobra.Command, args []string) {
	header, comps, err := completion.ProcessCommandArgs(cmd.Use, args)
	if unknownError, ok := err.(completion.UnknownResourceError); ok {
		logrus.Warnf("Unknown resource type: %s", unknownError)
		os.Exit(6)
	}
	if err != nil {
		logrus.Fatalf("Completion error: %s", err)
	}
	if len(comps) == 0 {
		logrus.Warn("No completion found")
		os.Exit(5)
	}
	fmt.Print(completion.FormatCompletion(header, comps))
}

func processResultFun(cmd *cobra.Command, args []string) {
	fzfResult := viper.GetString("fzf-result")
	sourceCmd := viper.GetString("source-cmd")
	res, err := completion.ProcessResult(fzfResult, sourceCmd)
	util.FatalIf(err)
	fmt.Print(res)
}

func statsFun(cmd *cobra.Command, args []string) {
	//res, err := completion.ProcessResult(fzfResult, sourceCmd)
	//util.FatalIf(err)
	//fmt.Print(res)
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

func addResultCmd(rootCmd *cobra.Command) {
	resultCmd := &cobra.Command{
		Use:     "process_result",
		Short:   "Process the result of the fzf output for the shell autocompletion. It will detect if namespace needs to be added or not and only output necessary fields.",
		Example: "kubectl-fzf-completion process_result --source-cmd \"get pods -l \" --fzf-result \"minikube kube-system tier=control-plane\"",
		Run:     processResultFun,
	}
	resultFlags := resultCmd.Flags()
	resultFlags.String("fzf-result", "", "Fzf output to process")
	resultFlags.String("source-cmd", "", "Initial completion command")
	err := viper.BindPFlags(resultFlags)
	util.FatalIf(err)
	rootCmd.AddCommand(resultCmd)
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
	addResultCmd(rootCmd)
	addStatsCmd(rootCmd)

	util.ConfigureViper()
	cobra.OnInitialize(util.ConfigureLog)
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Root command failed: %v", err)
	}
}
