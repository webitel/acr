package call

import (
	"fmt"
)

//set_user

func SetUser(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, prefix string

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			c.LogError("setUser", props, "name is require")
			return nil
		}

		name = fmt.Sprintf("%s@%s", name, c.Domain())

		prefix = getStringValueFromMap("prefix", props, "")
		if prefix != "" {
			name = fmt.Sprintf("%s %s", name, prefix)
		}

		err := c.Execute("set_user", name)
		if err != nil {
			c.LogError("setUser", name, err.Error())
			return err
		}
		c.LogDebug("setUser", name, "success")
	}

	return nil
}
