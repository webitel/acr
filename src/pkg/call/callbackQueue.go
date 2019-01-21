/**
 * Created by I. Navrotskyj on 30.10.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/logger"
)

func CallbackQueue(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var number, setVar string
	var err error
	var id = 0

	if props, ok = args.(map[string]interface{}); ok {
		number = c.ParseString(getStringValueFromMap("number", props, "${caller_id_number}"))
		err, id = c.acr.AddCallbackMember(
			c.Domain,
			getStringValueFromMap("queue", props, ""),
			number,
			getStringValueFromMap("widget", props, ""),
		)

		setVar = getStringValueFromMap("setVar", props, "")

		if err != nil {
			logger.Error("Call %s callbackQueue db error %s", c.Uuid, err.Error())
			if setVar != "" {
				return SetVar(c, setVar+"='"+err.Error()+"'")
			}
		} else {
			logger.Debug("Call %s callbackQueue add member (%d) %s successful", c.Uuid, id, number)
			if setVar != "" {
				return SetVar(c, fmt.Sprintf("%s=%d", setVar, id))
			}
		}
	} else {
		logger.Error("Call %s callbackQueue bad arguments %s", c.Uuid, args)

	}
	return nil
}
