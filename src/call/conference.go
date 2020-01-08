/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

func Conference(scope Scope, c *Call, args interface{}) error {

	var name, profile, data, pin, flag string
	var flags []interface{}
	var props map[string]interface{}
	var ok, oneElem bool

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			c.LogError("conference", props, "name is require")
			return nil
		}
		profile = getStringValueFromMap("profile", props, "default")
		pin = getStringValueFromMap("pin", props, "")

		data = name + "_" + c.Domain() + "@" + profile
		if pin != "" {
			data += "+" + pin
		}

		if _, ok = props["flags"]; ok {
			if flags, ok = props["flags"].([]interface{}); ok {

				data += "+flags{"
				for _, v := range flags {
					if flag, ok = v.(string); ok {
						if !oneElem {
							oneElem = true
							data += flag
						} else {
							data += "|" + flag
						}
					}
				}
				data += "}"
			}
		}

		err := c.Execute("conference", data)
		if err != nil {
			c.LogError("conference", data, err.Error())
			return err
		}

	} else {
		c.LogError("conference", args, "bad request")
	}
	return nil
}
