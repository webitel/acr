/**
 * Created by I. Navrotskyj on 08.11.17.
 */

package call

func SendEvent(scope Scope, c *Call, args interface{}) error {

	var props map[string]interface{}
	var ok bool

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("sendEvent", args, "bad request")
		return nil
	}

	m := make(map[string]string)
	m["Event-Name"] = "CUSTOM"
	//m["unique-id"] = c.Id()

	if getBoolValueFromMap("dump", props, false) {
		for k, v := range c.DumpVariables() {
			if isProtectedHeader(k) {
				continue
			}

			m[k] = v
		}
	}

	if _, ok = props["data"]; ok {
		if _, ok = props["data"].(map[string]interface{}); ok {
			data := props["data"].(map[string]interface{})
			for k, _ := range data {
				m[k] = c.ParseString(getStringValueFromMap(k, data, ""))
			}
		}
	}

	//TODO check \n or \r
	e := c.SendEvent(m, "ACR::HOOK")
	if e != nil {
		c.LogError("sendEvent", m, e.Error())
	} else {
		c.LogDebug("sendEvent", m, "success")
	}

	return nil
}

func isProtectedHeader(name string) bool {
	switch name {
	case "Event-Subclass", "variable_domain_name", "Event-Name", "Content-Type", "Reply-Text", "Presence-Call-Direction", "Core-UUID":
		return true
	}

	return false
}
