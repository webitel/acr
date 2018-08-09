package config

import (
	"flag"
	"github.com/webitel/acr/src/pkg/logger"
	. "github.com/webitel/gonfig"
	"os"
)

var Conf *Gonfig

func init() {
	filePath := flag.String("c", "./conf/config.json", "Config file path")
	flag.Parse()
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		logger.Error("No found config file: " + *filePath)
		os.Exit(1)
	}

	Conf = NewConfig(nil)
	Conf.Use("argv", NewEnvConfig(""))
	Conf.Use("env", NewEnvConfig(""))
	Conf.Use("local", NewJsonConfig(*filePath))
}
