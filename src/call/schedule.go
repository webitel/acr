/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"strconv"
	"strings"
)

const varCallbackCallExecute = "webitel_callback_call"

func Schedule(c *Call, args interface{}) error {
	var app, data, seconds, action, tmp string
	var ok bool
	var props map[string]interface{}

	if props, ok = args.(map[string]interface{}); ok {
		action = getStringValueFromMap("action", props, "")

		seconds = getStringValueFromMap("seconds", props, "2")
		if _, err := strconv.Atoi(seconds); err != nil {
			seconds = "2"
		}

		data = "+" + seconds + " "

		switch action {
		case "callback":
			return shedCallbackCall(c, seconds, props)

		case "hangup":
			app = "sched_hangup"
			data += getStringValueFromMap("data", props, "")

		case "goto":
			app = "sched_transfer"
			tmp = getStringValueFromMap("data", props, "")
			if tmp != "" {
				if strings.HasPrefix(tmp, "default:") {
					data += tmp[8:] + " XML default"
				} else if strings.HasPrefix(tmp, "public:") {
					data += tmp[7:] + " XML public"
				} else {
					data += tmp
				}
			}
		default:
			c.LogError("schedule", action, "bad action")
		}

		err := c.Execute(app, data)
		if err != nil {
			c.LogError("schedule", app+" "+data, err.Error())
			return err
		}
		c.LogDebug("schedule", app+" "+data, "success")
	} else {
		c.LogError("schedule", args, "bad request")
	}

	return nil
}

func shedCallbackCall(c *Call, seconds string, props map[string]interface{}) error {
	if c.Context() != model.CONTEXT_PUBLIC {
		c.LogError("schedule", props, "allow only public context")
		return nil
	}
	var number, destination string

	destination = c.ParseString(getStringValueFromMap("destination", props, c.Destination()))
	if destination == "" {
		c.LogError("schedule", props, "destination is require")
		return nil
	}

	if c.GetVariable(varCallbackCallExecute) == destination {
		c.LogDebug("schedule", props, "skip")
		return nil
	}

	number = c.ParseString(getStringValueFromMap("number", props, "${caller_id_number}"))
	if number == "" {
		c.LogError("schedule", props, "number is require")
		return nil
	}

	var s = "sched_api +%s none originate {^^:webitel_direction=callback:origination_caller_id_number='%s':hangup_after_bridge=true:domain_name='%s'" +
		":ignore_early_media=true:loopback_bowout_on_execute=false:loopback_bowout=false:%s='%s':caller_id_number='%s'" +
		":loopback_export=domain_name,caller_id_number,origination_caller_id_number,webitel_direction}loopback/%s/default %s  XML public %s %s"

	_, err := c.Api(fmt.Sprintf(s, seconds, c.Destination(), c.Domain(),
		varCallbackCallExecute, destination, number, number, destination, destination, destination))

	if err != nil {
		c.LogError("schedule", s, err.Error())
		return err
	}
	c.LogDebug("schedule", s, "success")

	return Hangup(c, getStringValueFromMap("cause", props, "REDIRECTION_TO_NEW_DESTINATION"))
}
