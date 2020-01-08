/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"strconv"
)

func Echo(scope Scope, c *Call, args interface{}) error {
	var err error
	var app = "echo"
	var data string

	switch args.(type) {
	case string:
		str := args.(string)
		if str != "" {
			app = "delay_echo"
			data = str
		}

	case int:
		app = "delay_echo"
		data = strconv.Itoa(args.(int))
	}

	err = c.Execute(app, data)
	if err != nil {
		c.LogError("echo", data, err.Error())
		return err
	}
	c.LogDebug("echo", data, "successful")
	return nil
}
