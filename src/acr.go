/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package main

import (
	"github.com/webitel/acr/src/pkg/acr"
	"github.com/webitel/acr/src/pkg/fs"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/logger"
	"net/http"
	_ "net/http/pprof"
	"fmt"
)




func main() {
	if config.Conf.Get("dev") == "true" {
		setDebug()
	}

	a := func(connection fs.Connection) {
		fmt.Println("OK a", connection)
		connection.Hangup("USER_BUSY")
	}
	b := func(connection fs.Connection) {
		fmt.Println("OK b", connection)
	}
	s := fs.NewEsl(":10030" ,a ,b)
	s.Listen()


	acr.New()
}

func setDebug() {
	//debug.SetGCPercent(-1)

	go func() {
		logger.Info("Start debug server on :8088")
		logger.Error("Debug: ", http.ListenAndServe(":8088", nil))
	}()

}
