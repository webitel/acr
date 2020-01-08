/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

func Script(scope Scope, c *Call, args interface{}) error {
	var name, typeScript, params, app string
	var props map[string]interface{}
	var ok bool
	var argsScript []interface{}

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			c.LogError("script", props, "name is require")
			return nil
		}

		typeScript = getStringValueFromMap("type", props, "lua")
		if typeScript == "js" {
			app = "js"
			params = "js/"
		} else {
			app = "lua"
			params = "lua/"
		}

		params += name

		if _, ok = props["parameters"]; ok {
			if argsScript, ok = props["parameters"].([]interface{}); ok {
				for _, v := range argsScript {
					if _, ok = v.(string); ok {
						params += ` "` + v.(string) + `"`
					}
				}
			}
		}

		err := c.Execute(app, params)
		if err != nil {
			c.LogError("script", app+" "+params, err.Error())
			return err
		}
		c.LogDebug("script", app+" "+params, "success")

	} else {
		c.LogError("script", args, "bad request")
	}

	return nil
}
