/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"regexp"
)

var replaceDomain = regexp.MustCompile(`@.*`)

func Agent(c *Call, args interface{}) error {
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
			logger.Error("Call %s agent name is require", c.Uuid)
			return nil
		}
		name += "@" + c.Domain

		status = getStringValueFromMap("status", props, "Available")

		_, err := c.Conn.Api("callcenter_config agent set status", name, status)
		if err != nil {
			logger.Error("Call %s agent set status error: %s", c.Uuid, err.Error())
			return err
		}

		state = getStringValueFromMap("state", props, "")
		if state != "" {
			_, err = c.Conn.Api("callcenter_config agent set state", name, state)
			if err != nil {
				logger.Error("Call %s agent set state error: %s", c.Uuid, err.Error())
				return err
			}
		}

		logger.Debug("Call %s agent %s successful", c.Uuid, name)

	} else {
		logger.Error("Call %s agent bad arguments %s", c.Uuid, args)
	}
	return nil
}
