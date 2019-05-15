/**
 * Created by I. Navrotskyj on 30.10.17.
 */

package call

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/webitel/acr/src/rpc"
)

func CDR(c *Call, args interface{}) error {
	var props, exportsVars, elastic map[string]interface{}
	var ok bool
	var err error
	var k string
	var v interface{}
	var request []byte

	if props, ok = args.(map[string]interface{}); ok {
		if exportsVars, ok = props["exportVar"].(map[string]interface{}); !ok {
			c.LogError("cdr", props, "exportVar is required")
			return nil
		}

		if elastic, ok = props["elastic"].(map[string]interface{}); !ok {
			c.LogError("cdr", props, "elastic is required")
			return nil
		}
		elastic["domain"] = c.Domain()

		request, err = json.Marshal(map[string]interface{}{
			"exec-api":  "elastic.request",
			"exec-args": elastic,
		})

		if err != nil {
			c.LogError("cdr", props, err.Error())
			return nil
		}

		err = c.router.app.FireRPCEventToStorage(
			"storage.commands.inbound",
			rpc.PublishingOption{
				Body:          []byte(c.ParseString(string(request))),
				ReplyTo:       c.router.app.GetRPCCommandsQueueName(),
				CorrelationId: c.Id(),
			},
		)

		if err != nil {
			c.LogError("cdr", args, err.Error())
			return err
		}

		d := c.router.app.AddRPCCommands(c.Id())
		c.LogDebug("cdr", string(d.Args), "response")
		if d.Args != nil {
			for k, v = range exportsVars {
				if _, ok = v.(string); ok {
					err = SetVar(c, "all:"+k+"="+gjson.GetBytes(d.Args, "data."+v.(string)).String()+"")
					if err != nil {
						c.LogError("cdr", map[string]interface{}{}, err.Error())
					}
				}
			}
		}

	}
	return nil
}
