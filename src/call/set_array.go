/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

func SetArray(c *Call, args interface{}) error {
	var tagName = "webitel_tags"
	var err error
	switch args.(type) {
	case string:
		return pushArray(c, tagName, args.(string))

	case []interface{}:
		return pushArray(c, tagName, args.([]interface{}))

	case map[string]interface{}:
		var val interface{}
		for tagName, val = range args.(map[string]interface{}) {
			err = pushArray(c, tagName, val)
			if err != nil {
				c.LogError("setArray", args, "bad request")
			}
		}
	default:
		c.LogError("setArray", args, "bad request")
		return nil
	}

	return nil
}

func pushArray(c *Call, tagName string, values interface{}) error {
	var err error
	switch values.(type) {
	case string:
		return pushArrayString(c, tagName, values.(string))
	case []interface{}:
		var ok bool
		for _, v := range values.([]interface{}) {
			if _, ok = v.(string); ok {
				err = pushArrayString(c, tagName, v.(string))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func pushArrayString(c *Call, tagName, val string) error {
	err := c.Execute("push", tagName+","+val)
	if err != nil {
		c.LogError("setArray", tagName+","+val, err.Error())
		return err
	}
	c.LogDebug("setArray", tagName+","+val, "success")
	return nil
}
