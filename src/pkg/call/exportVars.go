/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"encoding/json"
	"github.com/webitel/acr/src/pkg/logger"
)

func ExportVars(c *Call, args interface{}) error {

	if data, ok := args.([]interface{}); ok {
		vars := make(map[string]string)
		var v interface{}
		var tmp string
		for _, v = range data {
			if tmp, ok = v.(string); ok {
				vars[tmp] = c.Conn.ChannelData.Header.Get("variable_" + tmp)
			}
		}

		if len(vars) > 0 {
			body, err := json.Marshal(vars)
			if err != nil {
				logger.Error("Call %s exportVars to json error: %s", err.Error())
				return nil
			}
			err = SetVar(c, "all:webitel_data="+string(body))
			if err != nil {
				logger.Error("Call %s exportVars set webitel_data error: %s", err.Error())
				return err
			}

			err = SetVar(c, "cc_export_vars=webitel_data")
			if err != nil {
				logger.Error("Call %s exportVars set cc_export_vars error: %s", err.Error())
				return err
			}

			logger.Debug("Call %s exportVars: %s successful", c.Uuid, vars)
		}
	} else {
		logger.Error("Call %s exportVars bad arguments %s", c.Uuid, args)
	}
	return nil
}
