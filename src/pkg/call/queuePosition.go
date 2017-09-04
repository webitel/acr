/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"regexp"
	"strconv"
	"strings"
)

var positionTerms = regexp.MustCompile(`[^\r\n]+`)

func QueuePosition(c *Call, args interface{}) error {
	var props map[string]interface{}
	var varName, line string
	var ok bool
	var position int

	if c.CurrentQueue == "" {
		logger.Error("Call %s ccPosition current queue empty %s", c.Uuid, args)
	}

	if props, ok = args.(map[string]interface{}); ok {
		varName = getStringValueFromMap("var", props, "")
		if varName == "" {
			logger.Error("Call %s ccPosition var attribute is required %s", c.Uuid)
			return nil
		}

		data, err := c.Conn.Api("callcenter_config queue list members", c.CurrentQueue)
		if err != nil {
			logger.Error("Call %s ccPosition execute error: ", err.Error())
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
					if strings.Index(line, c.Uuid) != -1 {
						break
					}
				}
			}

			return SetVar(c, varName+"="+strconv.Itoa(position))
		}
	} else {
		logger.Error("Call %s ccPosition bad arguments %s", c.Uuid, args)
	}
	return nil
}
