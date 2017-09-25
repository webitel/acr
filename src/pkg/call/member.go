/**
 * Created by I. Navrotskyj on 25.09.17.
 */
package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Member(c *Call, args interface{}) error {
	var tmp string
	var props map[string]interface{}
	var ok bool
	var err error

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s member bad arguments %s", c.Uuid, args)
		return nil
	}

	tmp = getStringValueFromMap("id", props, "")

	if tmp != "" {
		delete(props, "id")
		err = c.acr.UpdateMember(tmp, props)
		tmp = "update"
	} else {
		err = c.acr.AddMember(props)
		tmp = "add"
	}

	if err != nil {
		logger.Error("Call %s: %s member: %s", c.Uuid, tmp, err.Error())
	} else {
		logger.Debug("Call %s: %s member successful", c.Uuid, tmp)
	}
	return nil
}
