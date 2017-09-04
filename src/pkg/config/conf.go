package config

import (
	. "github.com/Nomon/gonfig"
	"github.com/webitel/acr/src/pkg/logger"
	"os"
)

var Conf *Gonfig

func init() {
	filePath := "./conf/config.json"

	for i, s := range os.Args {
		if s == "-c" {
			if len(os.Args) > i+1 {
				filePath = os.Args[i+1]
			}
		}
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Error("No found config file: " + filePath)
		os.Exit(1)
	}

	Conf = NewConfig(nil)
	Conf.Use("argv", NewEnvConfig(""))
	Conf.Use("env", NewEnvConfig(""))
	Conf.Use("local", NewJsonConfig(filePath))
}
