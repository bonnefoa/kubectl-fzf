package util

import (
	"flag"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func ConfigureViper() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	viper.SetEnvPrefix("KUBECTL_FZF")
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.SetConfigName(".kubectl_fzf")
	viper.AddConfigPath("/etc/kubectl_fzf/")
	viper.AddConfigPath("$HOME")
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		FatalIf(err)
	}
}
