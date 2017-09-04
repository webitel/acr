/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"strconv"
	"strings"
)

func Schedule(c *Call, args interface{}) error {
	var app, data, seconds, action, tmp string
	var ok bool
	var props map[string]interface{}

	if props, ok = args.(map[string]interface{}); ok {
		action = getStringValueFromMap("action", props, "")

		seconds = getStringValueFromMap("seconds", props, "0")
		if _, err := strconv.Atoi(seconds); err != nil {
			seconds = "0"
		}

		data = "+" + seconds + " "

		switch action {
		case "hangup":
			app = "sched_hangup"
			data += getStringValueFromMap("data", props, "")

		case "goto":
			app = "sched_transfer"
			tmp = getStringValueFromMap("data", props, "")
			if tmp != "" {
				if strings.HasPrefix(tmp, "default:") {
					data += tmp[8:] + " XML default"
				} else if strings.HasPrefix(tmp, "public:") {
					data += tmp[7:] + " XML public"
				} else {
					data += tmp
				}
			}
		}

		_, err := c.SndMsg(app, data, true, false)
		if err != nil {
			logger.Error("Call %s schedule error: %s", c.Uuid, err.Error())
			return err
		}

	} else {
		logger.Error("Call %s schedule bad arguments %s", c.Uuid, args)
	}

	return nil
}
