/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var positionTerms = regexp.MustCompile(`[^\r\n]+`)

func QueuePosition(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var varName, line string
	var ok bool
	var position int

	if c.CurrentQueue == "" {
		c.LogError("queue_position", args, " queue is empty ")
		return nil
	}

	if props, ok = args.(map[string]interface{}); ok {
		varName = getStringValueFromMap("var", props, "")
		if varName == "" {
			c.LogError("queue_position", props, "var is require")
			return nil
		}

		data, err := c.Api(fmt.Sprintf("callcenter_config queue list members %s", c.CurrentQueue))
		if err != nil {
			c.LogError("queue_position", args, err.Error())
			return err
		}

		if len(data) > 0 {
			d := positionTerms.FindAllSubmatch(data, -1)
			for _, v := range d {
				if len(v) < 1 {
					continue
				}
				line = string(v[0])
				if strings.Index(line, "Trying") != -1 || strings.Index(line, "Waiting") != -1 {
					position++
					if strings.Index(line, c.Id()) != -1 {
						break
					}
				}
			}

			return c.SetVariable(varName + "=" + strconv.Itoa(position))
		}
	} else {
		c.LogError("queue_position", args, "bad request")
	}
	return nil
}
