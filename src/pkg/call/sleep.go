/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"strconv"
)

func Sleep(c *Call, args interface{}) error {
	var data = ""
	switch args.(type) {
	case string:
		data = c.ParseString(args.(string))
	case int:
		data = strconv.Itoa(args.(int))
	}

	logger.Debug("Call %s sleep %s", c.Uuid, data)
	_, err := c.SndMsg("sleep", data, true, true)
	if err != nil {
		logger.Error("Call %s sleep error: %s", c.Uuid, err.Error())
		return err
	}

	return nil
}
