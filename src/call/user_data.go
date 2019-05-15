/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"fmt"
)

func UserData(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, getVar, setVar string

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("userData", args, "bad request")
		return nil
	}

	if name = getStringValueFromMap("name", props, ""); name == "" {
		c.LogError("userData", args, "name is require")
		return nil
	}

	if getVar = getStringValueFromMap("var", props, ""); getVar == "" {
		c.LogError("userData", args, "var is require")
		return nil
	}

	if setVar = getStringValueFromMap("setVar", props, ""); setVar == "" {
		c.LogError("userData", args, "setVar is require")
		return nil
	}

	api := fmt.Sprintf("user_data %s@%s var %s", c.ParseString(name), c.Domain(), getVar)
	res, err := c.Api(api)
	if err != nil {
		c.LogError("userData", api, err.Error())
		return nil
	}

	getVar = string(res)
	c.LogDebug("userData", getVar, "success")
	if getVar == "-ERR no reply\n" {
		getVar = ""
	}

	return SetVar(c, setVar+"="+getVar)
}
