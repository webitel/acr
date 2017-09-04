/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"strconv"
)

func Echo(c *Call, args interface{}) error {
	var err error
	var app = "echo"
	var data string

	switch args.(type) {
	case string:
		str := args.(string)
		if str != "" {
			app = "delay_echo"
			data = str
		}

	case int:
		app = "delay_echo"
		data = strconv.Itoa(args.(int))
	}

	_, err = c.SndMsg(app, data, true, false)
	if err != nil {
		logger.Error("Call %s echo error: %s", c.Uuid, err.Error())
		return err
	}
	logger.Debug("Call %s %s successful", c.Uuid, app)
	return nil
}
