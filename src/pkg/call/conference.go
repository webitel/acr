/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Conference(c *Call, args interface{}) error {

	var name, profile, data, pin, flag string
	var flags []interface{}
	var props map[string]interface{}
	var ok, oneElem bool

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			logger.Error("Call %s conference name is required", c.Uuid)
			return nil
		}
		profile = getStringValueFromMap("profile", props, "default")
		pin = getStringValueFromMap("pin", props, "")

		data = name + "_" + c.Domain + "@" + profile
		if pin != "" {
			data += "+" + pin
		}

		if _, ok = props["flags"]; ok {
			if flags, ok = props["flags"].([]interface{}); ok {

				data += "+flags{"
				for _, v := range flags {
					if flag, ok = v.(string); ok {
						if !oneElem {
							oneElem = true
							data += flag
						} else {
							data += "|" + flag
						}
					}
				}
				data += "}"
			}
		}

		_, err := c.SndMsg("conference", data, true, true)
		if err != nil {
			logger.Error("Call %s log error: %s", c.Uuid, err.Error())
			return err
		}

	} else {
		logger.Error("Call %s conference bad arguments %s", c.Uuid, args)
	}
	return nil
}
