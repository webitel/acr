/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"fmt"
	"regexp"
)

var replaceDomain = regexp.MustCompile(`@.*`)

func Agent(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, state, status string

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			name = "${caller_id_number}"
		}

		name = c.ParseString(name)
		name = replaceDomain.ReplaceAllString(name, "")
		if name == "" {
			c.LogError("agent", args, "name is require")
			return nil
		}
		name += "@" + c.Domain()

		status = getStringValueFromMap("status", props, "Available")

		_, err := c.Api(fmt.Sprintf("callcenter_config agent set status %s '%s'", name, status))
		if err != nil {
			c.LogError("agent", name, err.Error())
			return err
		}

		state = getStringValueFromMap("state", props, "")
		if state != "" {
			_, err = c.Api(fmt.Sprintf("callcenter_config agent set state %s '%s'", name, state))
			if err != nil {
				c.LogError("agent", name, err.Error())
				return err
			}
		}
		c.LogDebug("agent", name, "successful")

	} else {
		c.LogError("agent", args, "bad request")
	}
	return nil
}
