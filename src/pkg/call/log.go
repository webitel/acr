/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Log(c *Call, args interface{}) error {

	if data, ok := args.(string); ok {
		_, err := c.SndMsg("log", "CONSOLE "+data, true, false)
		if err != nil {
			logger.Error("Call %s log error: %s", c.Uuid, err.Error())
			return err
		}
		logger.Debug("Call %s log %s successful", c.Uuid, data)
	} else {
		logger.Error("Call %s log bad arguments %s", c.Uuid, args)
	}
	return nil
}
