package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/logger"
)

//set_user

func SetUser(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, prefix string

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			logger.Error("Call %s setUser name is required", c.Uuid)
			return nil
		}

		name = fmt.Sprintf("%s@%s", name, c.Domain)

		prefix = getStringValueFromMap("prefix", props, "")
		if prefix != "" {
			name = fmt.Sprintf("%s %s", name, prefix)
		}

		_, err := c.SndMsg("set_user", name, true, false)
		if err != nil {
			logger.Error("Call %s setUser %s error: %s", c.Uuid, name, err.Error())
			return err
		}
		logger.Debug("Call %s setUser %s successful", c.Uuid, name)
	}

	return nil
}
