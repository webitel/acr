package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/logger"
)

func Cache(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var key, value string

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s cache bad arguments %v", c.Uuid, args)
		return nil
	}

	if key = getStringValueFromMap("key", props, ""); key == "" {
		logger.Error("Call %s cache key is required", c.Uuid)
		return nil
	}

	if value = getStringValueFromMap("value", props, ""); value != "" {
		value = c.ParseString(value)
	}

	switch getStringValueFromMap("method", props, "") {
	case "add":
		c.acr.AddToDomainCache(c, key, value, int64(getIntValueFromMap("expire", props, 60)))
		logger.Debug("Call %s add to cache %v = %v", c.Uuid, key, value)
		break
	case "get":
		exportVarName := getStringValueFromMap("setVar", props, "")
		if exportVarName == "" {
			logger.Error("Call %s get cache setVar is required", c.Uuid)
			return nil
		}
		value, ok = c.acr.GetFromDomainCache(c, key)
		if ok {
			return SetVar(c, fmt.Sprintf("%s=%s", exportVarName, value))
		} else {
			logger.Debug("Call %s not found cache %v", c.Uuid, key)
		}

		break
	case "remove":
		c.acr.RemoveFromDomainCache(c, key)
		logger.Debug("Call %s remove cache %v", c.Uuid, key)
		break
	}

	return nil
}
