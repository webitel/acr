/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

//todo need test
func SipRedirect(c *Call, args interface{}) error {
	var data, tmp string
	var v interface{}
	var ok bool

	switch args.(type) {
	case string:
		data = args.(string)
	case []interface{}:
		for _, v = range args.([]interface{}) {
			if _, ok = v.(string); ok {
				data += "," + v.(string)
			}
		}
		if len(data) > 1 {
			data = data[1:]
		}
	}

	if data == "" {
		logger.Error("Call %s sipRedirect bad arguments: %v", c.Uuid, args)
		return nil
	}

	tmp = c.GetChannelVar("Caller-Channel-Answered-Time")

	if tmp == "0" || tmp == "" {
		tmp = "redirect"
	} else {
		tmp = "deflect"
	}

	_, err := c.SndMsg(tmp, data, true, true)
	if err != nil {
		logger.Error("Call %s sipRedirect error: %s", c.Uuid, err.Error())
		return err
	}
	logger.Debug("Call %s sipRedirect %s -> %s successful", c.Uuid, tmp, data)

	return nil
}
