/**
 * Created by I. Navrotskyj on 30.10.17.
 */

package call

import (
	"fmt"
)

func CallbackQueue(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var number, setVar, comment string
	var err error
	var id = 0

	if props, ok = args.(map[string]interface{}); ok {
		number = c.ParseString(getStringValueFromMap("number", props, "${caller_id_number}"))
		result := <-c.router.app.Store.CallbackQueue().CreateMember(
			c.Domain(),
			c.ParseString(getStringValueFromMap("queue", props, "")),
			number,
			c.ParseString(getStringValueFromMap("widget", props, "")),
		)
		setVar = getStringValueFromMap("setVar", props, "")

		if result.Err != nil {
			c.LogError("callback", props, result.Err.Error())
			if setVar != "" {
				return SetVar(scope, c, setVar+"='"+err.Error()+"'")
			}
			return nil
		} else {
			c.LogDebug("callback", map[string]interface{}{"Id": result.Data}, "successful")
			comment = getStringValueFromMap("comment", props, "")
			if comment != "" {
				c.router.app.Store.CallbackQueue().CreateMemberComment(result.Data.(int64), c.Domain(), "ACR", c.ParseString(comment))
			}
			if setVar != "" {
				return SetVar(scope, c, fmt.Sprintf("%s=%d", setVar, id))
			}
		}
	} else {
		c.LogError("callback", args, "bad request")
	}
	return nil
}
