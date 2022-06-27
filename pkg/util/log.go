package util

import (
	"flag"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func SetLogConfFlags() {
	flag.String("log-level", "info", "Log level to use")
}

type LogConf struct {
	LogLevel logrus.Level
}

func GetLogConf() LogConf {
	logLevelStr := viper.GetString("log-level")
	logLevel, err := logrus.ParseLevel(logLevelStr)
	FatalIf(err)

	l := LogConf{}
	l.LogLevel = logLevel
	return l
}

func ConfigureLog() {
	logConf := GetLogConf()
	logrus.Debugf("Setting log level %v", logConf.LogLevel)
	logrus.SetLevel(logConf.LogLevel)
}
