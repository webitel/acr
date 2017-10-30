/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"encoding/json"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/rpc"
)

//err = c.acr.FireRPCEvent([]byte("IGOR"), "*.message.system")
//if err != nil {
//		logger.Error("Call %s fire rpc error: %s", c.Uuid, err.Error())
//}

func Log(c *Call, args interface{}) error {

	if data, ok := args.(string); ok {

		if c.IsDebugLog() {
			msgJson := make(map[string]interface{})
			msgJson["action"] = "log"
			msgJson["domain"] = c.Domain
			msgJson["message"] = c.ParseString(data)
			if body, err := json.Marshal(msgJson); err == nil {
				c.acr.FireRPCEventToEngine("*.broadcast.message."+c.GetRouteId(), rpc.PublishingOption{
					Body: body,
				})
			} else {
				logger.Error("Call %s log marshal json message error: %s", c.Uuid, err.Error())
			}
		}

		return sendLogToFs(c, data)
	} else {
		logger.Error("Call %s log bad arguments %s", c.Uuid, args)
	}
	return nil
}

func sendLogToFs(c *Call, data string) error {
	_, err := c.SndMsg("log", "CONSOLE "+data, true, false)
	if err != nil {
		logger.Error("Call %s log error: %s", c.Uuid, err.Error())
		return err
	}
	logger.Debug("Call %s log %s successful", c.Uuid, data)
	return nil
}
