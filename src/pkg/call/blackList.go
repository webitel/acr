/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/router"
	"gopkg.in/mgo.v2/bson"
	"regexp"
)

var replaceNumber = regexp.MustCompile(`\D`)

func BlackList(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, varName, number string
	var actions []interface{}
	var count int
	var err error

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if name == "" {
			logger.Error("Call %s blackList name is require", c.Uuid)
			return nil
		}

		varName = getStringValueFromMap("variable", props, "caller_id_number")
		number = c.ParseString("${" + varName + "}")

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
				actions, ok = props["actions"].([]interface{})
			}

			if len(actions) == 0 {
				actions = []interface{}{
					bson.M{
						"hangup": "INCOMING_CALL_BARRED",
					},
					bson.M{
						"break": true,
					},
				}
			} else {
				actions = append(actions, bson.M{
					"break": true,
				})
			}

			iterator := router.NewIterator(actions, c)
			routeCallIterator(c, iterator)
			logger.Debug("Call %s blackList number %s bared", c.Uuid, number)
		} else {
			logger.Debug("Call %s blackList skip number %s", c.Uuid, number)
		}

	} else {
		logger.Error("Call %s blackList bad arguments %s", c.Uuid, args)
	}

	return nil
}
