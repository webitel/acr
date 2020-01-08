/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"strings"
)

//TODO ADD old goto support - delete new version;
func GoTo(scope Scope, c *Call, args interface{}) error {
	var tag string
	var ok bool

	if tag, ok = args.(string); !ok || tag == "" {
		c.LogError("goto", args, "bad request")
		return nil
	}

	if strings.HasPrefix(tag, "local:") {
		if scope.Goto(c.ParseString(tag[6:])) {
			c.LogDebug("goto", tag, "start")
		} else {
			c.LogWarn("goto", tag, "not found")
		}

		return nil
	} else if strings.HasPrefix(tag, "default:") {
		tag = tag[8:] + " XML default"
	} else if strings.HasPrefix(tag, "public:") {
		tag = tag[7:] + " XML public"
	} else if strings.HasPrefix(tag, "socket:") {
		tag = tag[7:]
		err := c.Execute("socket", tag)
		if err != nil {
			c.LogError("goto", tag, err.Error())
			return err
		}
		c.SetBreak()
		//TODO close socket
		return nil

	} else {
		c.LogError("goto", args, "bad request")
		return nil
	}

	err := c.Execute("transfer", tag)
	if err != nil {
		c.LogError("goto", tag, err.Error())
		return err
	}
	c.SetBreak()
	c.LogDebug("goto", tag, "success")

	return nil
}
