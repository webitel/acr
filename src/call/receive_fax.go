/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

func ReceiveFax(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var err error
	var vars []string
	var email string

	if c.GetVariable("Caller-Channel-Answered-Time") == "0" {
		err = Answer(scope, c, "")
		if err != nil {
			return err
		}
	}

	err = c.Execute("playback", "silence_stream://2000")
	if err != nil {
		c.LogError("receiveFax", nil, err.Error())
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
		c.LogError("receiveFax", props, err.Error())
		return err
	}

	err = c.Execute("rxfax", "/recordings/${uuid}.tif")
	if err != nil {
		c.LogError("receiveFax", props, err.Error())
		return err
	}
	c.LogDebug("receiveFax", props, "success")

	return nil
}
