/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

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
				logger.Error("Call %s setArray bad arguments %s", c.Uuid, val)
			}
		}
	default:
		logger.Error("Call %s setArray bad arguments %s", c.Uuid, args)
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
	_, err := c.SndMsg("push", tagName+","+val, true, true)
	if err != nil {
		logger.Error("Call %s pushArray error: %s", c.Uuid, err.Error())
		return err
	}
	logger.Debug("Call %s pushArray %s = %s successful", c.Uuid, tagName, val)
	return nil
}
