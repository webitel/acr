/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

func VoiceMail(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool

	var check, announce, auth, user string
	var setVars []interface{}
	var err error

	if props, ok = args.(map[string]interface{}); ok {
		check = getStringValueFromMap("check", props, "")
		user = getStringValueFromMap("user", props, "")

		if check == "true" {
			announce = getStringValueFromMap("announce", props, "")
			if announce != "" {
				setVars = append(setVars, "vm_announce_cid="+announce)
			}

			auth = getStringValueFromMap("auth", props, "")
			switch auth {
			case "true":
				setVars = append(setVars, "voicemail_authorized=false")
			case "false":
				setVars = append(setVars, "voicemail_authorized=true")
			default:
				setVars = append(setVars, "voicemail_authorized=${sip_authorized}")
			}

			err = SetVar(c, setVars)
			if err != nil {
				return err
			}

			err = c.Execute("voicemail", "check default "+c.Domain()+" "+user)
			if err != nil {
				c.LogError("voicemail", "check default "+c.Domain()+" "+user, err.Error())
				return err
			}
			c.LogDebug("voicemail", "check default "+c.Domain()+" "+user, "success")
		} else {
			if user == "" {
				c.LogError("voicemail", props, "name is require")
				return nil
			}

			if getStringValueFromMap("skip_greeting", props, "") == "true" {
				setVars = append(setVars, "skip_greeting=true")
			}

			if getStringValueFromMap("skip_instructions", props, "") == "true" {
				setVars = append(setVars, "skip_instructions=true")
			}

			if _, ok = props["cc"]; ok {
				cc := parseCC(props["cc"], c.Domain())
				if cc != "" {
					setVars = append(setVars, "vm_cc="+cc)
				}
			}

			if len(setVars) > 0 {
				err = SetVar(c, setVars)
				if err != nil {
					c.LogError("voicemail", props, err.Error())
					return err
				}
			}

			api := "default " + c.Domain() + " " + user
			err = c.Execute("voicemail", api)
			if err != nil {
				c.LogError("voicemail", api, err.Error())
				return err
			}
			c.LogDebug("voicemail", api, "success")
		}
	} else {
		c.LogError("voicemail", args, "bad request")
	}

	return nil
}

func parseCC(parameters interface{}, domainName string) string {
	var ok bool

	switch parameters.(type) {
	case string:
		return parameters.(string) // + "@" + domainName

	case []interface{}:
		var cc = ""
		for _, v := range parameters.([]interface{}) {
			if _, ok = v.(string); ok {
				cc += "," + v.(string) // + "@" + domainName
			}
		}
		if len(cc) > 0 {
			return cc[1:]
		}
	}
	return ""
}
