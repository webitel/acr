/**
 * Created by I. Navrotskyj on 26.09.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Limit(c *Call, args interface{}) error {
	var name, varName string

	switch args.(type) {
	case string:
		name = args.(string)
	case map[string]interface{}:
		prop := args.(map[string]interface{})
		name = getStringValueFromMap("name", prop, "")
		varName = getStringValueFromMap("setVar", prop, "")
	}

	if name == "" {
		logger.Error("Call %s bad limit parameters: %v", c.Uuid, args)
		return nil
	}

	_, err := c.SndMsg("limit", "hash "+c.Domain+" "+name, true, true)
	if err != nil {
		logger.Error("Call %s limit error: %s", c.Uuid, err)
		return err
	}
	logger.Debug("Call %s limit: %s successful", c.Uuid, name)

	if varName != "" {
		return SetVar(c, varName+"="+c.GetChannelVar("variable_limit_usage_"+c.Domain+"_"+name))
	}
	return nil
}
