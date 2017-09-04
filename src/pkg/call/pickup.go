/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

//todo need test
func Pickup(c *Call, args interface{}) error {

	if data, ok := args.(string); ok && data != "" {
		_, err := c.SndMsg("pickup", data+"@"+c.Domain, true, true)
		if err != nil {
			logger.Error("Call %s pickup error: %s", c.Uuid, err.Error())
			return err
		}
		logger.Debug("Call %s pickup %s successful", c.Uuid, data)
	} else {
		logger.Error("Call %s pickup bad arguments %s", c.Uuid, args)
	}
	return nil
}
