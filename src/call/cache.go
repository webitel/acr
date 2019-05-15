package call

import (
	"fmt"
)

func Cache(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var key, value string

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("cache", args, "bad request")
		return nil
	}

	if key = getStringValueFromMap("key", props, ""); key == "" {
		c.LogError("cache", args, "key is required")
		return nil
	}

	if value = getStringValueFromMap("value", props, ""); value != "" {
		value = c.ParseString(value)
	}

	switch getStringValueFromMap("method", props, "") {
	case "add":
		c.router.AddToDomainCache(c, key, value, int64(getIntValueFromMap("expire", props, 60)))
		c.LogDebug("cache", fmt.Sprintf("add %s=%s", key, value), "successful")
		break
	case "get":
		exportVarName := getStringValueFromMap("setVar", props, "")
		if exportVarName == "" {
			c.LogError("cache", props, "setVar is required")
			return nil
		}
		value, ok = c.router.GetFromDomainCache(c, key)
		if ok {
			c.LogDebug("cache", key, value)
			return SetVar(c, fmt.Sprintf("%s=%s", exportVarName, value))
		} else {
			c.LogDebug("cache", key, "not found")
		}

		break
	case "remove":
		c.router.RemoveFromDomainCache(c, key)
		c.LogDebug("cache", key, "remove successful")
		break
	}

	return nil
}
