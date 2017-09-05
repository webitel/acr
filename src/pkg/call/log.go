/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"encoding/json"
	"github.com/webitel/acr/src/pkg/logger"
)

//err = c.acr.FireRPCEvent([]byte("IGOR"), "*.message.system")
//if err != nil {
//		logger.Error("Call %s fire rpc error: %s", c.Uuid, err.Error())
//}

func Log(c *Call, args interface{}) error {

	if data, ok := args.(string); ok {
		return sendLogToFs(c, data)
	} else if props, ok := args.(map[string]interface{}); ok {
		msgText := getStringValueFromMap("message", props, "")
		if _, ok = props["to"]; ok {
			msgJson := make(map[string]interface{})
			msgJson["to"] = props["to"]
			msgJson["message"] = msgText
			msgJson["callId"] = c.Uuid
			if body, err := json.Marshal(msgJson); err == nil {
				c.acr.FireRPCEvent(body, "*.message.system")
			} else {
				logger.Error("Call %s log marshal json message error: %s", c.Uuid, err.Error())
			}

		}
		return sendLogToFs(c, msgText)
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
