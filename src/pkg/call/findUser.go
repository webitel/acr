/**
 * Created by I. Navrotskyj on 13.12.18.
 */

package call

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/rpc"
)

func FindUser(c *Call, args interface{}) error {
	var props, exportsVars map[string]interface{}
	var ok bool
	var err error
	var request []byte
	variables := make([]string, 1, 1)

	if props, ok = args.(map[string]interface{}); ok {
		if exportsVars, ok = props["exportVariables"].(map[string]interface{}); !ok {
			logger.Error("Call %s findUser exportVariables is required", c.Uuid)
			return nil
		}

		requestArgs := map[string]interface{}{
			"domain": c.Domain,
		}

		if _, ok = props["filter"]; ok {
			requestArgs["filter"] = props["filter"]
		}

		if _, ok = props["columns"]; ok {
			requestArgs["columns"] = props["columns"]
		}

		request, err = json.Marshal(map[string]interface{}{
			"exec-api":  "account.user_data",
			"exec-args": requestArgs,
		})

		if err != nil {
			logger.Error("Call %s findUser marshal request error: %s", c.Uuid, err.Error())
			return nil
		}

		c.acr.FireRPCEventToEngine("rpc", rpc.PublishingOption{
			Body:          []byte(c.ParseString(string(request))),
			ReplyTo:       c.acr.GetRPCCommandsQueueName(),
			CorrelationId: c.Uuid,
		})

		response := c.acr.AddRPCCommands(c.Uuid)
		logger.Debug("Call %s findUser receive response", c.Uuid)

		if response.Args != nil {
			for k, v := range exportsVars {
				if _, ok = v.(string); ok {
					variables = append(variables, k+"="+gjson.GetBytes(response.Args, "data."+v.(string)).String())
				}
			}

			return SetVar(c, variables)
		}
	}

	return nil
}
