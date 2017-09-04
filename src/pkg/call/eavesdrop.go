/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

//todo need test spy=false
func Eavesdrop(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var app, data, user string
	var err error

	if props, ok = args.(map[string]interface{}); ok {
		if user = getStringValueFromMap("user", props, ""); user == "" {
			logger.Error("Call %s eavesdrop user is required", c.Uuid)
			return nil
		}

		if c.GetChannelVar("Caller-Channel-Answered-Time") == "0" {
			err = Answer(c, "")
			if err != nil {
				logger.Error("Call %s eavesdrop answer error: %v", c.Uuid, err.Error())
				return err
			}
		}

		err = SetVar(c, "webitel_direction=eavesdrop")
		if err != nil {
			logger.Error("Call %s eavesdrop set direction error: %v", c.Uuid, err.Error())
			return err
		}

		if user == "all" {
			err = SetVar(c, "eavesdrop_require_group="+c.Domain)
			if err != nil {
				logger.Error("Call %s eavesdrop set eavesdrop_require_group error: %v", c.Uuid, err.Error())
				return err
			}

			_, err = c.SndMsg("eavesdrop", c.Domain, true, true)
			if err != nil {
				logger.Error("Call %s all eavesdrop error: %s", c.Uuid, err.Error())
				return err
			}

		} else {
			user = c.ParseString(user)
			if _, ok = props["spy"]; ok {
				if _, ok = props["spy"].(bool); ok {
					if props["spy"].(bool) {
						ok = true
					} else {
						ok = false
					}
				}
			}
			if ok {
				app = "userspy"
				data = user + "@${domain_name}"
			} else {
				app = "eavesdrop"
				data = "${hash(select/spymap/${domain_name}-" + user + ")}"
			}

			_, err = c.SndMsg(app, data, true, true)
			if err != nil {
				logger.Error("Call %s %s error: %s", c.Uuid, app, err.Error())
				return err
			}
		}

	} else {
		logger.Error("Call %s eavesdrop bad arguments %s", c.Uuid, args)
	}

	logger.Debug("Call %s eavesdrop successful", c.Uuid)
	return nil
}
