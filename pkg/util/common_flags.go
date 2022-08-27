package util

import (
	"flag"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetCommonCliFlags(fs *pflag.FlagSet, defaultLogLevel string) {
	fs.String("log-level", defaultLogLevel, "Log level to use")
	fs.Bool("cpu-profile", false, "Start with cpu profiling")
}

func ConfigureViper() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	err := viper.BindPFlags(pflag.CommandLine)
	FatalIf(err)

	viper.SetEnvPrefix("KUBECTL_FZF")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.SetConfigName(".kubectl_fzf")
	viper.AddConfigPath("/etc/kubectl_fzf/")
	viper.AddConfigPath("$HOME")
	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		FatalIf(err)
	}
}
