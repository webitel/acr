/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Sleep(c *Call, args interface{}) error {

	logger.Debug("Call %s sleep %v", c.Uuid, args)
	_, err := c.SndMsg("sleep", parseInterfaceToString(args), true, true)
	if err != nil {
		logger.Error("Call %s sleep error: %s", c.Uuid, err.Error())
		return err
	}

	return nil
}
