/**
 * Created by I. Navrotskyj on 04.09.17.
 */

package call

import (
	"github.com/tidwall/gjson"
	"github.com/webitel/acr/src/pkg/logger"
	"strconv"
)

func STT(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var maxSec, varName string
	var err error

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s stt bad arguments %s", c.Uuid, args)
		return nil
	}

	maxSec = getStringValueFromMap("maxSec", props, "10")

	if i, err := strconv.Atoi(maxSec); err != nil || i > 15 {
		maxSec = "10"
	}

	var rate = ""
	if rate = c.GetChannelVar("read_rate"); rate == "" {
		rate = "8000"
	}

	varName = getStringValueFromMap("setVar", props, "stt_response")

	var data = "http_cache://$${cdr_url}/sys/stt?rate=" + rate +
		"&lang=" + getStringValueFromMap("lang", props, "en-US") +
		"&setVar=" + varName +
		"&key=" + getStringValueFromMap("key", props, "") +
		"&callId=" + c.Uuid +
		"&reply=" + c.acr.GetRPCCommandsQueueName() +
		"&type=.wav " + maxSec + " " + getStringValueFromMap("silenceThresh", props, "200") + " " + getStringValueFromMap("silenceHits", props, "3")

	_, err = c.SndMsg("record", data, true, false)
	if err != nil {
		logger.Error("Call %s record error: %s", c.Uuid, err.Error())
		return err
	}

	d := c.acr.AddRPCCommands(c.Uuid)

	SetVar(c, varName+"="+gjson.GetBytes(d.Args, "stt.result.0.alternative.0.transcript").String())
	return nil
}
