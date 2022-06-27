package main

import (
	"fmt"
	"kubectlfzf/pkg/completion"
	"kubectlfzf/pkg/util"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func fallbackComp(cmd *cobra.Command, args []string) {
	logrus.Debugf("Fallback comp %s", args)
	os.Exit(1)
}

func completeFun(cmd *cobra.Command, args []string) {
	comps, err := completion.ProcessCommandArgs(args)
	if err != nil {
		logrus.Errorf("Completion error: %s", err)
	}
	fmt.Print(strings.Join(comps, ""))
}

func resultFun(cmd *cobra.Command, args []string) {
	fzfResult := viper.GetString("fzf-result")
	sourceCmd := viper.GetString("initial-cmd")
	logrus.Debugf("Processing fzf result %s", fzfResult)
	logrus.Debugf("Source command %s", sourceCmd)
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "kubectl_completion",
		Run: fallbackComp,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	verbs := []string{"get", "label", "describe", "delete", "annotate", "edit"}
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
		Use: "result",
		Run: resultFun,
	}
	resultCmd.Flags().String("fzf-result", "", "Fzf output to process")
	resultCmd.Flags().String("initial-cmd", "", "Initial completion command")
	rootCmd.AddCommand(resultCmd)

	util.SetClusterConfFlags()
	util.SetLogConfFlags()
	util.ConfigureViper()
	cobra.OnInitialize(util.ConfigureLog)

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Root command failed: %v", err)
	}
}
