/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package main

import (
	"github.com/webitel/acr/src/app"
	"github.com/webitel/acr/src/call"
	"github.com/webitel/acr/src/config"
	"github.com/webitel/wlog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	if config.Conf.Get("dev") == "true" {
		setDebug()
	}

	interruptChan := make(chan os.Signal, 1)

	acr := app.New()
	defer acr.Shutdown()

	router := call.InitCallRouter(acr)

	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

	router.Stop()

}

func setDebug() {
	//debug.SetGCPercent(1)

	go func() {
		wlog.Info("Start debug server on :8088")
		http.ListenAndServe(":8088", nil)
	}()

}
