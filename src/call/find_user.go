/**
 * Created by I. Navrotskyj on 13.12.18.
 */

package call

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/webitel/acr/src/rpc"
)

func FindUser(scope Scope, c *Call, args interface{}) error {
	var props, exportsVars map[string]interface{}
	var ok bool
	var err error
	var request []byte
	variables := make([]string, 1, 1)

	if props, ok = args.(map[string]interface{}); ok {
		if exportsVars, ok = props["exportVariables"].(map[string]interface{}); !ok {
			c.LogError("findUser", props, "exportVariables is required")
			return nil
		}

		requestArgs := map[string]interface{}{
			"domain": c.Domain(),
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
			c.LogError("findUser", requestArgs, err.Error())
			return nil
		}

		c.router.app.FireRPCEventToEngine("rpc", rpc.PublishingOption{
			Body:          []byte(c.ParseString(string(request))),
			ReplyTo:       c.router.app.GetRPCCommandsQueueName(),
			CorrelationId: c.Id(),
		})

		response := c.router.app.AddRPCCommands(c.Id())
		c.LogDebug("findUser", string(response.Args), "successful")

		if response.Args != nil {
			for k, v := range exportsVars {
				if _, ok = v.(string); ok {
					variables = append(variables, k+"="+gjson.GetBytes(response.Args, "data."+v.(string)).String())
				}
			}

			return SetVar(scope, c, variables)
		}
	}

	return nil
}
