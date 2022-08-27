package util

import (
	"fmt"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LogConf struct {
	LogLevel logrus.Level
}

func getLogConf() LogConf {
	logLevelStr := viper.GetString("log-level")
	logLevel, err := logrus.ParseLevel(logLevelStr)
	FatalIf(err)

	l := LogConf{}
	l.LogLevel = logLevel
	return l
}

func configureLog() {
	logConf := getLogConf()
	logrus.Debugf("Setting log level %v", logConf.LogLevel)
	logrus.SetLevel(logConf.LogLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return f.Function + "()", fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		},
		FullTimestamp: true,
	})
}
