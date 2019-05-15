/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"fmt"
)

//todo need test spy=false
func Eavesdrop(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var app, data, user string
	var err error

	if props, ok = args.(map[string]interface{}); ok {
		if user = getStringValueFromMap("user", props, ""); user == "" {
			c.LogError("eavesdrop", props, "user is require")
			return nil
		}

		if c.GetVariable("Caller-Channel-Answered-Time") == "0" {
			err = Answer(c, "")
			if err != nil {
				c.LogError("eavesdrop", props, "answer error: "+err.Error())
				return err
			}
		}

		err = SetVar(c, "webitel_direction=eavesdrop")
		if err != nil {
			c.LogError("eavesdrop", props, "set direction error: "+err.Error())
			return err
		}

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

		user = fmt.Sprintf("%s@%s", user, c.Domain())

		result := <-c.router.app.Store.Call().GetIdByPresence(user)

		if result.Err != nil {
			c.LogError("eavesdrop", args, "store error: "+result.Err.Error())
			return nil
		}
		presenceId := result.Data.(string)

		if ok {
			app = "userspy"
			data = fmt.Sprintf("%s %s", user, presenceId)
		} else {
			app = "eavesdrop"
			data = presenceId
		}

		err = c.Execute(app, data)
		if err != nil {
			c.LogError("eavesdrop", data, err.Error())
			return err
		}

	} else {
		c.LogError("eavesdrop", args, "bad request")
	}

	c.LogDebug("eavesdrop", args, "successful")
	return nil
}
