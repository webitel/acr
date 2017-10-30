/**
 * Created by I. Navrotskyj on 30.10.17.
 */

package call

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/rpc"
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
			logger.Error("Call %s CDR exportVar is required", c.Uuid)
			return nil
		}

		if elastic, ok = props["elastic"].(map[string]interface{}); !ok {
			logger.Error("Call %s CDR elastic bad parameters: %v", c.Uuid, props["elastic"])
			return nil
		}
		elastic["domain"] = c.Domain

		request, err = json.Marshal(map[string]interface{}{
			"exec-api":  "elastic.request",
			"exec-args": elastic,
		})

		if err != nil {
			logger.Error("Call %s CDR marshal parameters: %s", c.Uuid, err.Error())
			return nil
		}

		err = c.acr.FireRPCEventToStorage(
			"storage.commands.inbound",
			rpc.PublishingOption{
				Body:          []byte(c.ParseString(string(request))),
				ReplyTo:       c.acr.GetRPCCommandsQueueName(),
				CorrelationId: c.Uuid,
			},
		)

		if err != nil {
			logger.Error("Call %s CDR RPC error %s", c.Uuid, err.Error())
			return err
		}

		d := c.acr.AddRPCCommands(c.Uuid)
		if d.Args != nil {
			for k, v = range exportsVars {
				if _, ok = v.(string); ok {
					err = SetVar(c, "all:"+k+"="+gjson.GetBytes(d.Args, "data."+v.(string)).String()+"")
					if err != nil {
						logger.Error("Call %s CDR setVat error: %s", c.Uuid, err.Error())
					}
				}
			}
		}

	}
	return nil
}
