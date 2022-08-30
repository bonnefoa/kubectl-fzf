package httpserver

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type HttpServerConfigCli struct {
	ListenAddress string
	Debug         bool
}

func SetHttpServerConfigFlags(fs *pflag.FlagSet) {
	fs.String("listen-address", ":8080", "Listen address of the http server")
	fs.Bool("http-debug", false, "Activate debug mode of the http server")
}

func GetHttpServerConfigCli() HttpServerConfigCli {
	h := HttpServerConfigCli{}
	h.ListenAddress = viper.GetString("listen-address")
	h.Debug = viper.GetBool("http-debug")
	return h
}
