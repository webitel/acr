/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/logger"
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
			logger.Error("Call %s bad schedule action: ", c.Uuid, action)
		}

		_, err := c.SndMsg(app, data, true, false)
		if err != nil {
			logger.Error("Call %s schedule error: %s", c.Uuid, err.Error())
			return err
		}

	} else {
		logger.Error("Call %s schedule bad arguments %s", c.Uuid, args)
	}

	return nil
}

func shedCallbackCall(c *Call, seconds string, props map[string]interface{}) error {
	if c.context != CONTEXT_PUBLIC {
		logger.Error("Call %s schedule callback error: only public context", c.Uuid)
		return nil
	}
	var number, destination string

	destination = c.ParseString(getStringValueFromMap("destination", props, c.DestinationNumber))
	if destination == "" {
		logger.Error("Call %s schedule callback error: not found destination", c.Uuid)
		return nil
	}

	if c.GetChannelVar(varCallbackCallExecute) == destination {
		logger.Debug("Call %s schedule callback skip destination %s", c.Uuid, destination)
		return nil
	}

	number = c.ParseString(getStringValueFromMap("number", props, "${caller_id_number}"))
	if number == "" {
		logger.Error("Call %s schedule callback error: not found number", c.Uuid)
		return nil
	}

	logger.Debug("Call %s schedule callback try call %s to %s", c.Uuid, number, destination)

	var s = "sched_api +%s none originate {^^:webitel_direction=callback:origination_caller_id_number='%s':hangup_after_bridge=true:domain_name='%s'" +
		":ignore_early_media=true:loopback_bowout_on_execute=false:loopback_bowout=false:%s='%s':caller_id_number='%s'" +
		":loopback_export=domain_name,caller_id_number,origination_caller_id_number,webitel_direction}loopback/%s/default %s  XML public %s %s"

	_, err := c.Conn.Api(fmt.Sprintf(s, seconds, c.DestinationNumber, c.Domain, varCallbackCallExecute, destination, number, number, destination, destination, destination))

	if err != nil {
		logger.Error("Call %s schedule callback error : ", c.Uuid, err.Error())
		return err
	}

	return Hangup(c, getStringValueFromMap("cause", props, "REDIRECTION_TO_NEW_DESTINATION"))
}
