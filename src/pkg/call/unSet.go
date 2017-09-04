/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func UnSet(c *Call, args interface{}) error {

	if data, ok := args.(string); ok {
		_, err := c.SndMsg("unset", data, true, true)
		if err != nil {
			logger.Error("Call %s unSet error: %s", c.Uuid, err.Error())
			return err
		}
		logger.Debug("Call %s unSet %s successful", c.Uuid, data)
	} else {
		logger.Error("Call %s unSet bad arguments %s", c.Uuid, args)
	}
	return nil
}
