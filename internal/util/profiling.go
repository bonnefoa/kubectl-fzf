package util

import (
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func DoMemoryProfile() {
	memProfile := viper.GetString("mem-profile")
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			logrus.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			logrus.Fatal("could not write memory profile: ", err)
		}
	}
}
