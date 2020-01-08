/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

//todo need test
func SipRedirect(scope Scope, c *Call, args interface{}) error {
	var data, tmp string
	var v interface{}
	var ok bool

	switch args.(type) {
	case string:
		data = args.(string)
	case []interface{}:
		for _, v = range args.([]interface{}) {
			if _, ok = v.(string); ok {
				data += "," + v.(string)
			}
		}
		if len(data) > 1 {
			data = data[1:]
		}
	}

	if data == "" {
		c.LogError("sipRedirect", args, "bad request")
		return nil
	}

	tmp = c.GetVariable("Caller-Channel-Answered-Time")

	if tmp == "0" || tmp == "" {
		tmp = "redirect"
	} else {
		tmp = "deflect"
	}

	err := c.Execute(tmp, data)
	if err != nil {
		c.LogError("sipRedirect", tmp+" "+data, err.Error())
		return err
	}
	c.LogDebug("sipRedirect", tmp+" "+data, "success")
	return nil
}
