package util

import (
	"flag"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetCommonCliFlags(fs *pflag.FlagSet, defaultLogLevel string) {
	fs.String("log-level", defaultLogLevel, "Log level to use")
	fs.String("cpu-profile", "", "Destination file for cpu profiling")
}

func CommonInitialization() {
	configureLog()
	cpuProfile := viper.GetString("cpu-profile")
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			logrus.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

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
