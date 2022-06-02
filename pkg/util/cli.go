package util

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func ParseFlags() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName(".kubectl_fzf")
	viper.AddConfigPath("/etc/kubectl_fzf/")
	viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		FatalIf(err)
	}
}
