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
	"sync/atomic"
)


func main() {
	if config.Conf.Get("dev") == "true" {
		setDebug()
	}

	var i int64 = 0

	a := func(connection fs.Connection) {
		atomic.AddInt64(&i, 1)
		connection.Execute("hangup", "USER_BUSY")
	}
	b := func(connection fs.Connection) {
		atomic.AddInt64(&i, -1)
		//if connection.GetVar("Event-Name") != "CHANNEL_HANGUP_COMPLETE" {
		//	panic(connection.GetVar("Event-Name"))
		//}
		fmt.Println("count: ", i)
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
