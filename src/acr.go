/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package main

import (
	"github.com/webitel/acr/src/pkg/acr"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/logger"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	if config.Conf.Get("dev") == "true" {
		setDebug()
	}
	acr.New()
}

func setDebug() {
	//debug.SetGCPercent(-1)

	go func() {
		logger.Info("Start debug server on :8088")
		logger.Error("Debug: ", http.ListenAndServe(":8088", nil))
	}()

}
