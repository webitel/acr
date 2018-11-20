/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
	"github.com/webitel/acr/src/pkg/router"
	"regexp"
)

var replaceNumber = regexp.MustCompile(`\D`)

func BlackList(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, varName, number string
	var actions models.ArrayApplications
	var count int
	var err error

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			logger.Error("Call %s blackList name is require", c.Uuid)
			return nil
		}

		varName = getStringValueFromMap("number", props, "caller_id_number")
		number = c.ParseString(varName)

		if number == "" {
			logger.Error("Call %s blackList number is require", c.Uuid)
			return nil
		}

		number = replaceNumber.ReplaceAllString(number, "")

		err, count = c.acr.CheckBlackList(c.Domain, name, number)
		if err != nil {
			logger.Error("Call %s blackList CheckBlackList error: ", c.Uuid, err.Error())
			return err
		}

		if count > 0 {
			if _, ok = props["actions"]; ok {
				if _, ok = props["actions"].([]interface{}); ok {
					actions = router.ArrInterfaceToArrayApplication(props["actions"].([]interface{}))
				}
			}

			if len(actions) == 0 {
				actions = models.ArrayApplications{
					map[string]interface{}{
						"hangup": "INCOMING_CALL_BARRED",
					},
					map[string]interface{}{
						"break": true,
					},
				}
			}

			iterator := router.NewIterator(actions, c)
			routeCallIterator(c, iterator)

			if c.GetBreak() {
				logger.Debug("Call %s break from blacklist", c.Uuid)
			}

			logger.Debug("Call %s blackList number %s bared", c.Uuid, number)
		} else {
			logger.Debug("Call %s blackList skip number %s", c.Uuid, number)
		}

	} else {
		logger.Error("Call %s blackList bad arguments %s", c.Uuid, args)
	}

	return nil
}
