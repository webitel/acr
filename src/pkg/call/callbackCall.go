/**
 * Created by I. Navrotskyj on 13.11.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/logger"
)

const varCallbackCallExecute = "webitel_callback_call"

func CallbackCall(c *Call, args interface{}) error {
	if c.context != CONTEXT_PUBLIC {
		logger.Error("Call %s callbackCall error: only public context", c.Uuid)
		return nil
	}

	var props map[string]interface{}
	var ok bool
	var number, destination string

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s callbackCall bad arguments %v", c.Uuid, args)
		return nil
	}

	destination = getStringValueFromMap("destination", props, c.DestinationNumber)
	if destination == "" {
		logger.Error("Call %s callbackCall error: not found destination", c.Uuid)
		return nil
	}

	if c.GetChannelVar(varCallbackCallExecute) == destination {
		logger.Debug("Call %s callbackCall skip destination %s", c.Uuid, destination)
		return nil
	}

	number = c.ParseString(getStringValueFromMap("number", props, "${caller_id_number}"))
	if number == "" {
		logger.Error("Call %s callbackCall error: not found number", c.Uuid)
		return nil
	}

	logger.Debug("Call %s callbackCall try call %s to %s", c.Uuid, number, destination)

	var s = "api_hangup_hook=originate {^^:webitel_direction=callback:origination_caller_id_number='%s':hangup_after_bridge=true:domain_name='%s'" +
		":ignore_early_media=true:loopback_bowout_on_execute=false:loopback_bowout=false:%s='%s':caller_id_number='%s'" +
		":loopback_export=domain_name,caller_id_number,origination_caller_id_number,webitel_direction}loopback/%s/default %s  XML public %s %s"

	err := SetVar(c, fmt.Sprintf(s, c.DestinationNumber, c.Domain, varCallbackCallExecute, destination, number, number, destination, destination, destination))

	if err != nil {
		logger.Error("Call %s callbackCall error set api_reporting_hook: ", c.Uuid, err.Error())
		return err
	}

	return Hangup(c, "REDIRECTION_TO_NEW_DESTINATION")
}
