/**
 * Created by I. Navrotskyj on 04.09.17.
 */

package call

import (
	"github.com/tidwall/gjson"
	"strconv"
)

func STT(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var maxSec, varName string
	var err error

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("stt", args, "bad request")
		return nil
	}

	maxSec = getStringValueFromMap("maxSec", props, "10")

	if i, err := strconv.Atoi(maxSec); err != nil || i > 15 {
		maxSec = "10"
	}

	var rate = ""
	if rate = c.GetVariable("read_rate"); rate == "" {
		rate = "8000"
	}

	varName = getStringValueFromMap("setVar", props, "stt_response")

	var data = "http_cache://$${cdr_url}/sys/stt?rate=" + rate +
		"&lang=" + getStringValueFromMap("lang", props, "en-US") +
		"&setVar=" + varName +
		"&key=" + getStringValueFromMap("key", props, "") +
		"&callId=" + c.Id() +
		"&reply=" + c.router.app.GetRPCCommandsQueueName() +
		"&type=.wav " + maxSec + " " + getStringValueFromMap("silenceThresh", props, "200") + " " + getStringValueFromMap("silenceHits", props, "3")

	err = c.Execute("record", data)
	if err != nil {
		c.LogError("stt", data, err.Error())
		return err
	}

	d := c.router.app.AddRPCCommands(c.Id())
	result := gjson.GetBytes(d.Args, "stt.results.0.alternatives.0.transcript").String()
	c.LogDebug("stt", result, "success")
	c.SetVariable(varName + "=" + result)
	return nil
}
