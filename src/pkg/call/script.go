/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Script(c *Call, args interface{}) error {
	var name, typeScript, params, app string
	var props map[string]interface{}
	var ok bool
	var argsScript []interface{}

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			logger.Error("Call %s script name is required", c.Uuid)
			return nil
		}

		typeScript = getStringValueFromMap("type", props, "lua")
		if typeScript == "js" {
			app = "js"
			params = "js/"
		} else {
			app = "lua"
			params = "lua/"
		}

		params += name

		if _, ok = props["parameters"]; ok {
			if argsScript, ok = props["parameters"].([]interface{}); ok {
				for _, v := range argsScript {
					if _, ok = v.(string); ok {
						params += ` "` + v.(string) + `"`
					}
				}
			}
		}

		_, err := c.SndMsg(app, params, true, false)
		if err != nil {
			logger.Error("Call %s script %s error: %s", c.Uuid, name, err.Error())
			return err
		}
		logger.Debug("Call %s script %s successful", c.Uuid, name)

	} else {
		logger.Error("Call %s script bad arguments %s", c.Uuid, args)
	}

	return nil
}
