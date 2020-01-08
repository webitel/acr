/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/router"
	"regexp"
)

var replaceNumber = regexp.MustCompile(`\D`)

func BlackList(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, varName, number string
	var actions model.ArrayApplications
	//var err error

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			c.LogError("blackList", args, "name is required")
			return nil
		}

		varName = getStringValueFromMap("number", props, "${caller_id_number}")
		number = c.ParseString(varName)

		if number == "" {
			c.LogError("blackList", args, "number is required")
			return nil
		}

		number = replaceNumber.ReplaceAllString(number, "")

		result := <-c.router.app.Store.BlackList().CountNumbers(c.Domain(), name, number)
		if result.Err != nil {
			c.LogError("blackList", args, result.Err.Error())
			return result.Err
		}

		if result.Data.(int) > 0 {
			if _, ok = props["actions"]; ok {
				if _, ok = props["actions"].([]interface{}); ok {
					actions = router.ArrInterfaceToArrayApplication(props["actions"].([]interface{}))
				}
			}

			if len(actions) == 0 {
				actions = model.ArrayApplications{
					map[string]interface{}{
						"hangup": "INCOMING_CALL_BARRED",
					},
					map[string]interface{}{
						"break": true,
					},
				}
			}

			iterator := router.NewIterator("BlackList", actions, c)
			c.iterateCallApplication(iterator)
			c.LogDebug("blackList", number, "bared")
		} else {
			c.LogDebug("blackList", number, "skip")
		}

	} else {
		c.LogError("blackList", args, "bad request")
	}

	return nil
}
