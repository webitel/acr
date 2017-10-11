/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
)

func RingBack(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok, all bool
	var err error
	var elem models.Application

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

		logger.Debug("Call %s ringback successful", c.Uuid)

	} else {
		logger.Error("Call %s ringback bad arguments: %v", c.Uuid, args)
		return nil
	}
	return nil
}

func setRingBack(c *Call, elem models.Application, varName string, all bool) error {
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
		false,
		getStringValueFromMap("lang", elem, ""),
		getStringValueFromMap("method", elem, ""),
	)
	err = SetVar(c, data)
	if err != nil {
		logger.Error("Call %s ringback set %s error: %s", c.Uuid, varName, err.Error())
		return err
	}
	return nil
}
