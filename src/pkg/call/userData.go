/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func UserData(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, getVar, setVar string

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s userData bad arguments %v", c.Uuid, args)
		return nil
	}

	if name = getStringValueFromMap("name", props, ""); name == "" {
		logger.Error("Call %s userData name is required", c.Uuid)
		return nil
	}

	if getVar = getStringValueFromMap("var", props, ""); getVar == "" {
		logger.Error("Call %s userData var is required", c.Uuid)
		return nil
	}

	if setVar = getStringValueFromMap("setVar", props, ""); setVar == "" {
		logger.Error("Call %s userData var is required", c.Uuid)
		return nil
	}

	res, err := c.Conn.Api("user_data", c.ParseString(name)+"@"+c.Domain, "var", getVar)
	if err != nil {
		logger.Error("Call %s userData api error %s", c.Uuid, err.Error())
		return nil
	}

	getVar = string(res)
	logger.Debug("Call %s userData api response %s", c.Uuid, getVar)
	if getVar == "-ERR no reply\n" {
		getVar = ""
	}

	return SetVar(c, setVar+"="+getVar)
}
