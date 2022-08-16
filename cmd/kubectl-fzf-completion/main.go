package main

import (
	"fmt"
	"kubectlfzf/pkg/completion"
	"kubectlfzf/pkg/k8s/fetcher"
	"kubectlfzf/pkg/util"
	"os"
	"strings"

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
	fmt.Print(header, strings.Join(comps, "\n"))
}

func processResultFun(cmd *cobra.Command, args []string) {
	fzfResult := viper.GetString("fzf-result")
	sourceCmd := viper.GetString("source-cmd")
	res, err := completion.ProcessResult(fzfResult, sourceCmd)
	util.FatalIf(err)
	fmt.Print(res)
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "kubectl_fzf_completion",
	}
	rootFlags := rootCmd.PersistentFlags()
	util.SetCommonCliFlags(rootFlags)
	err := viper.BindPFlags(rootFlags)
	util.FatalIf(err)

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
		rootCmd.AddCommand(cmd)
	}

	resultCmd := &cobra.Command{
		Use: "process_result",
		Run: processResultFun,
	}
	resultFlags := resultCmd.Flags()
	resultFlags.String("fzf-result", "", "Fzf output to process")
	resultFlags.String("source-cmd", "", "Initial completion command")
	fetcher.SetFetchConfigFlags(resultFlags)
	err = viper.BindPFlags(resultFlags)
	util.FatalIf(err)
	rootCmd.AddCommand(resultCmd)

	util.ConfigureViper()
	cobra.OnInitialize(util.ConfigureLog)
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Root command failed: %v", err)
	}
}
