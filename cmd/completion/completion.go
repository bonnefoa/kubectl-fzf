package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// import (
// 	"os"
//
// 	"k8s.io/cli-runtime/pkg/genericclioptions"
// 	"k8s.io/component-base/cli"
// 	"k8s.io/kubectl/pkg/cmd"
//
// 	// Import to initialize client auth plugins.
// 	_ "k8s.io/client-go/plugin/pkg/client/auth"
// 	"k8s.io/kubectl/pkg/cmd/get"
// 	"k8s.io/kubectl/pkg/cmd/plugin"
// )
//
// func main() {
// 	command := cmd.NewDefaultKubectlCommand()
// 	var defaultConfigFlags = genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)
// 	o := cmd.KubectlOptions{
// 		PluginHandler: cmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes),
// 		Arguments:     os.Args,
// 		ConfigFlags:   defaultConfigFlags,
// 		IOStreams:     genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
// 	}
// 	kubeConfigFlags := o.ConfigFlags
// 	if kubeConfigFlags == nil {
// 		kubeConfigFlags = defaultConfigFlags
// 	}
// 	flags := cmds.PersistentFlags()
// 	flags.BoolVar(&warningsAsErrors, "warnings-as-errors", warningsAsErrors, "Treat warnings received from the server as errors and exit with a non-zero exit code")
//
// 	kubeConfigFlags.AddFlags(flags)
// 	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
// 	matchVersionKubeConfigFlags.AddFlags(flags)
// 	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
//
// 	getCmd := get.NewCmdGet("kubectl", f)
//
// 	code := cli.Run(command)
// 	os.Exit(code)
// }

func CompGetResourceList(cmd *cobra.Command, toComplete string) []string {
	var comps []string

	return comps
}

func completeGetFun(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var comps []string
	if len(args) == 0 {
		comps = CompGetResourceList(cmd, toComplete)
	} else {
		return comps, cobra.ShellCompDirectiveNoFileComp
		//comps = CompGetResource(f, cmd, args[0], toComplete)
		//if len(args) > 1 {
		//comps = cmdutil.Difference(comps, args[1:])
		//}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func main() {
	var rootCmd = &cobra.Command{
		Use:               "hugo",
		Short:             "Hugo is a very fast static site generator",
		ValidArgsFunction: completeGetFun,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
