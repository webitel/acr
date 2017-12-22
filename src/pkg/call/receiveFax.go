/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func ReceiveFax(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var err error
	var vars []string
	var email string

	if c.GetChannelVar("Caller-Channel-Answered-Time") == "0" {
		err = Answer(c, "")
		if err != nil {
			logger.Error("Call %s receiveFax answer error: %v", c.Uuid, err.Error())
			return err
		}
	}

	_, err = c.SndMsg("playback", "silence_stream://2000", true, true)
	if err != nil {
		logger.Error("Call %s receiveFax playback error: %s", c.Uuid, err.Error())
		return err
	}

	if props, ok = args.(map[string]interface{}); ok {
		if _, ok = props["enable_t38"]; ok {
			if _, ok = props["enable_t38"].(bool); ok {
				if props["enable_t38"].(bool) {
					ok = true
				} else {
					ok = false
				}
			}
		}

		if ok {
			vars = []string{
				"fax_enable_t38_request=true",
				"fax_enable_t38=true",
			}
		}

		if _, ok = props["email"]; ok {
			email = parseEmail(props["email"])
			if email == "none" {
				email = ""
			}
		}
	}

	vars = append(vars, "execute_on_fax_success=lua FaxUpload.lua ${uuid} ${domain_name} "+email,
		"execute_on_fax_failure=system /bin/rm /recordings/${uuid}.tif")

	err = multiSetVar(c, vars)
	if err != nil {
		logger.Error("Call %s receiveFax set vars error: %s", c.Uuid, err.Error())
		return err
	}

	_, err = c.SndMsg("rxfax", "/recordings/${uuid}.tif", true, true)
	if err != nil {
		logger.Error("Call %s receiveFax error: %s", c.Uuid, err.Error())
		return err
	}
	logger.Debug("Call %s receiveFax successful", c.Uuid)

	return nil
}
