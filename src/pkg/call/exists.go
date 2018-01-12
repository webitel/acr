package call

import (
	"bytes"
	"fmt"
	"github.com/webitel/acr/src/pkg/logger"
)

func (c *Call) ExistsResource(resource string, props map[string]interface{}) bool {
	return exists(c, resource, props)
}

func Exists(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var varName, resource string

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s exists bad arguments %v", c.Uuid, args)
		return nil
	}

	varName = getStringValueFromMap("setVar", props, "")
	resource = getStringValueFromMap("resource", props, "")
	if varName == "" || resource == "" {
		logger.Error("Call %s exists setVar or resource is required", c.Uuid)
		return nil
	}

	if exists(c, resource, props) {
		return SetVar(c, varName+"=true")
	}

	return SetVar(c, varName+"=false")
}

func exists(c *Call, resource string, props map[string]interface{}) bool {
	switch resource {
	case "media":
		return existsMedia(c, getStringValueFromMap("name", props, ""), getStringValueFromMap("type", props, ""))
	case "dialer":
		return existsDialer(c, getStringValueFromMap("name", props, ""))
	case "account":
		return ExistsAccount(c, getStringValueFromMap("name", props, ""))
	case "queue":
		return existsQueue(c, getStringValueFromMap("name", props, ""))
	}
	return false
}

func existsMedia(c *Call, name, typeFile string) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}
	return c.acr.ExistsMediaFile(name, typeFile, c.Domain)
}

func existsDialer(c *Call, name string) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}
	return c.acr.ExistsDialer(name, c.Domain)
}

func existsQueue(c *Call, name string) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}
	return c.acr.ExistsQueue(name, c.Domain)
}

var bytesTrueString = []byte("true")

func ExistsAccount(c *Call, name string) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}

	res, err := c.Conn.Api(fmt.Sprintf("user_exists id %s %s", name, c.Domain))
	if err != nil {
		logger.Error("Call %s existsAccount error: %s", c.Uuid, err.Error())
		return false
	}

	return bytes.Equal(res, bytesTrueString)
}
