/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"encoding/json"
	"github.com/webitel/acr/src/rpc"
	"strconv"
)

//err = c.acr.FireRPCEvent([]byte("IGOR"), "*.message.system")
//if err != nil {
//		logger.Error("Call %s fire rpc error: %s", c.Uuid, err.Error())
//}

func Log(scope Scope, c *Call, args interface{}) error {

	if data, ok := args.(string); ok {

		if c.IsDebugLog() {
			msgJson := make(map[string]interface{})
			msgJson["action"] = "log"
			msgJson["domain"] = c.Domain()
			msgJson["message"] = c.ParseString(data)
			if body, err := json.Marshal(msgJson); err == nil {
				c.router.app.FireRPCEventToEngine("*.broadcast.message."+strconv.Itoa(c.RouteId()), rpc.PublishingOption{
					Body: body,
				})
			} else {
				c.LogError("log", data, err.Error())
			}
		}

		return sendLogToFs(c, data)
	} else {
		c.LogError("log", args, "bad request")
	}
	return nil
}

func sendLogToFs(c *Call, data string) error {
	err := c.Execute("log", "CONSOLE "+data)
	if err != nil {
		c.LogError("log", data, err.Error())
		return err
	}
	c.LogDebug("log", data, "successful")
	return nil
}
