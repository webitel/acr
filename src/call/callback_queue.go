/**
 * Created by I. Navrotskyj on 30.10.17.
 */

package call

import (
	"encoding/json"
	"fmt"
)

func CallbackQueue(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var number, setVar, comment string
	var id = 0

	if props, ok = args.(map[string]interface{}); ok {
		number = c.ParseString(getStringValueFromMap("number", props, "${caller_id_number}"))
		member, err := c.router.app.CreateCallbackMember(
			c.Domain(),
			c.ParseString(getStringValueFromMap("queue", props, "")),
			number,
			c.ParseString(getStringValueFromMap("widget", props, "")),
		)
		setVar = getStringValueFromMap("setVar", props, "")

		if err != nil {
			c.LogError("callback", props, err.Error())
			if setVar != "" {
				return SetVar(c, setVar+"='"+err.Error()+"'")
			}
			return nil
		} else {
			c.LogDebug("callback", map[string]interface{}{"Id": member.Id}, "successful")
			m := map[string]interface{}{
				"webitel_hook_retries": 0,
				"Event-Name":           "CUSTOM",
				"Event-Subclass":       "engine::callback_member_add",
				"domain":               c.Domain(),
				"variable_domain_name": c.Domain(),
				"id":                   member.Id,
				"number":               member.Number,
				"queue_id":             member.QueueId,
				"queue_name":           member.QueueName,
				"widget_id":            member.WidgetId,
				"widget_name":          member.WidgetName,
				"created_on":           member.CreatedOn,
			}

			if body, e := json.Marshal(m); e == nil {
				c.router.app.FireRPCEventToHook(body)
			}

			comment = getStringValueFromMap("comment", props, "")
			if comment != "" {
				result := <-c.router.app.Store.CallbackQueue().CreateMemberComment(member.Id, c.Domain(), "ACR", c.ParseString(comment))
				if result.Err != nil {
					c.LogError("callback", args, result.Err.Error())
				}
				var i int
				i, ok = result.Data.(int)
				if i == 0 {
					c.LogError("callback_comment", args, "no insert callback comment")
				} else {
					c.LogDebug("callback_comment", args, fmt.Sprintf("store new comment %d for member %d", i, member.Id))
				}
			}
			if setVar != "" {
				return SetVar(c, fmt.Sprintf("%s=%d", setVar, id))
			}
		}
	} else {
		c.LogError("callback", args, "bad request")
	}
	return nil
}
