/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"strings"
)

func Park(c *Call, args interface{}) error {
	var data map[string]interface{}
	var ok bool
	var name, lot, auto string

	if data, ok = args.(map[string]interface{}); ok {
		if _, ok = data["name"]; !ok {
			logger.Error("Call %s park name is required", c.Uuid)
			return nil
		}

		if _, ok = data["lot"]; !ok {
			logger.Error("Call %s park lot is required", c.Uuid)
			return nil
		}

		if name, ok = data["name"].(string); !ok || name == "" {
			logger.Error("Call %s park name must string", c.Uuid)
			return nil
		}

		if lot, ok = data["lot"].(string); !ok || lot == "" {
			logger.Error("Call %s park lot must string", c.Uuid)
			return nil
		}

		if _, ok = data["auto"]; ok {
			if auto, ok = data["auto"].(string); ok {
				auto = "auto " + auto + " "
				lot = strings.Replace(lot, "-", " ", -1)
			}
		}

		str := name + "@${domain_name} " + auto + lot
		_, err := c.SndMsg("valet_park", str, true, true)
		if err != nil {
			logger.Error("Call %s park error: %s", c.Uuid, err)
			return err
		}
		logger.Debug("Call %s park: %s successful", c.Uuid, str)
	} else {
		logger.Error("Call %s park bad arguments %s", c.Uuid, args)
	}
	return nil
}
