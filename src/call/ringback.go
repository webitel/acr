/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/model"
)

func RingBack(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok, all bool
	var err error
	var elem model.Application

	if props, ok = args.(map[string]interface{}); ok {

		if _, ok = props["all"]; ok {
			if _, ok = props["all"].(bool); ok {
				if props["all"].(bool) {
					all = true
				}
			}
		}

		if _, ok = props["call"]; ok {
			if elem, ok = props["call"].(map[string]interface{}); ok {
				err = setRingBack(c, elem, "ringback", all)
				if err != nil {
					return err
				}
			}
		}

		if _, ok = props["hold"]; ok {
			if elem, ok = props["hold"].(map[string]interface{}); ok {
				err = setRingBack(c, elem, "hold_music", all)
				if err != nil {
					return err
				}
			}
		}

		if _, ok = props["transfer"]; ok {
			if elem, ok = props["transfer"].(map[string]interface{}); ok {
				err = setRingBack(c, elem, "transfer_ringback", all)
				if err != nil {
					return err
				}
			}
		}
		c.LogDebug("ringback", props, "success")
	} else {
		c.LogError("ringback", args, "bad request")
		return nil
	}
	return nil
}

func setRingBack(c *Call, elem model.Application, varName string, all bool) error {
	var err error
	var data string

	if all || getBoolValueFromMap("all", elem, false) {
		data = "all:"
	}
	data += varName + "=" + getPlaybackFileString(
		c,
		getStringValueFromMap("type", elem, ""),
		getStringValueFromMap("name", elem, ""),
		getBoolValueFromMap("refresh", elem, false),
		true,
		getStringValueFromMap("lang", elem, ""),
		getStringValueFromMap("method", elem, ""),
	)
	err = c.SetVariable(data)
	if err != nil {
		return err
	}
	return nil
}
