/**
 * Created by I. Navrotskyj on 26.09.17.
 */

package call

func Limit(c *Call, args interface{}) error {
	var name, varName string

	switch args.(type) {
	case string:
		name = args.(string)
	case map[string]interface{}:
		prop := args.(map[string]interface{})
		name = getStringValueFromMap("name", prop, "")
		varName = getStringValueFromMap("setVar", prop, "")
	}

	if name == "" {
		c.LogError("limit", args, "name is required")
		return nil
	}

	err := c.Execute("limit", "hash "+c.Domain()+" "+name)
	if err != nil {
		c.LogError("limit", "hash "+c.Domain()+" "+name, err.Error())
		return err
	}
	c.LogDebug("limit", "hash "+c.Domain()+" "+name, "success")

	if varName != "" {
		return SetVar(c, varName+"="+c.GetVariable("variable_limit_usage_"+c.Domain()+"_"+name))
	}
	return nil
}
