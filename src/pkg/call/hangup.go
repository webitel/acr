/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Hangup(c *Call, args interface{}) error {
	switch args.(type) {
	case string:
		return hangupChannel(c, args.(string))
	case map[string]interface{}:
		params := args.(map[string]interface{})
		uuid := c.ParseString(getStringValueFromMap("uuid", params, ""))
		if uuid == "" {
			logger.Error("Call %s bad hangup parameters: uuid is required")
			return nil
		}
		return hangupByUuid(
			c,
			uuid,
			c.ParseString(getStringValueFromMap("cause", params, "")),
		)

	default:
		logger.Error("Call %s hangup cause must string or object", c.Uuid)
		return nil
	}
	return nil
}

func hangupChannel(c *Call, cause string) error {
	_, err := c.SndMsg("hangup", cause, true, true)
	if err != nil {
		logger.Error("Call %s hangup error: %s", c.Uuid, err.Error())
		return err
	}
	c.Conn.Disconnected = true
	logger.Debug("Call %s hangup cause: %s successful", c.Uuid, cause)
	return nil
}

func hangupByUuid(c *Call, uuid, cause string) error {
	_, err := c.Conn.Api("uuid_kill " + uuid + " " + cause)
	if err != nil {
		logger.Error("Call %s hangup by uuid error: %s", c.Uuid, cause)
	}
	logger.Debug("Call %s kill call uuid %s by cause %s", c.Uuid, uuid, cause)
	return nil
}
