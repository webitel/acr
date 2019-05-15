package call

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/webitel/acr/src/model"
)

func (c *Call) ExistsResource(resource string, props map[string]interface{}) bool {
	return exists(c, resource, props)
}

func Exists(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var varName, resource string

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("exists", args, "bad request")
		return nil
	}

	varName = getStringValueFromMap("setVar", props, "")
	resource = getStringValueFromMap("resource", props, "")
	if varName == "" || resource == "" {
		c.LogError("exists", args, "setVar or resource is required")
		return nil
	}

	if exists(c, resource, props) {
		c.LogDebug("exists", resource, "true")
		return SetVar(c, varName+"=true")
	}
	c.LogDebug("exists", resource, "false")
	return SetVar(c, varName+"=false")
}

func exists(c *Call, resource string, props map[string]interface{}) bool {
	switch resource {
	case "media":
		return existsMedia(c, getStringValueFromMap("name", props, ""), getStringValueFromMap("type", props, ""))
	case "dialer":
		return existsDialer(c, getStringValueFromMap("name", props, ""), props["member"])
	case "account":
		return ExistsAccount(c, getStringValueFromMap("name", props, ""))
	case "queue":
		return existsQueue(c, getStringValueFromMap("name", props, ""))
	case "callback":
		return existsCallbackQueue(c, getStringValueFromMap("name", props, ""))
	}
	return false
}

func existsMedia(c *Call, name, typeFile string) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}
	result := <-c.router.app.Store.Media().ExistsFile(name, typeFile, c.Domain())
	if result.Err != nil {
		c.LogError("exists", name, result.Err.Error())
		return false
	}
	return result.Data.(bool)
}

func existsDialer(c *Call, name string, member interface{}) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}

	if member == nil {
		result := <-c.router.app.Store.OutboundQueue().Exists(name, c.Domain())
		if result.Err != nil {
			c.LogError("exists", name, result.Err.Error())
			return false
		}
		return result.Data.(bool)
	} else {
		var r *model.OutboundQueueExistsMemberRequest
		body, err := json.Marshal(member)
		if err != nil {
			c.LogError("exists", member, err.Error())
			return false
		}
		if err = json.Unmarshal([]byte(c.ParseString(string(body))), &r); err != nil {
			return false
		}
		result := <-c.router.app.Store.OutboundQueue().ExistsMember(name, c.Domain(), r)
		if result.Err != nil {
			c.LogError("exists", member, err.Error())
			return false
		}
		return result.Data.(bool)
	}
}

func existsQueue(c *Call, name string) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}
	result := <-c.router.app.Store.InboundQueue().Exists(c.Domain(), name)
	if result.Err != nil {
		c.LogError("exists", name, result.Err.Error())
		return false
	}
	return result.Data.(bool)
}

func existsCallbackQueue(c *Call, name string) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}
	result := <-c.router.app.Store.CallbackQueue().Exists(c.Domain(), name)
	if result.Err != nil {
		c.LogError("exists", name, result.Err.Error())
		return false
	}
	return result.Data.(bool)
}

var bytesTrueString = []byte("true")

func ExistsAccount(c *Call, name string) bool {
	name = c.ParseString(name)
	if name == "" {
		return false
	}

	res, err := c.Api(fmt.Sprintf("user_exists id %s %s", name, c.Domain()))
	if err != nil {
		c.LogError("exists", name, err.Error())
		return false
	}

	return bytes.Equal(res, bytesTrueString)
}
