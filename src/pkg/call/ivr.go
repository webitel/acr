/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func IVR(c *Call, args interface{}) error {

	if data, ok := args.(string); ok {
		_, err := c.SndMsg("ivr", data+"@"+c.Domain, true, true)
		if err != nil {
			logger.Error("Call %s ivr error: %s", c.Uuid, err.Error())
			return err
		}
		logger.Debug("Call %s ivr %s successful", c.Uuid, data)
	} else {
		logger.Error("Call %s ivr bad arguments %s", c.Uuid, args)
	}
	return nil
}
