/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Hangup(c *Call, args interface{}) error {
	var cause = ""
	switch args.(type) {
	case string:
		cause = args.(string)
	default:
		logger.Error("Call %s hangup cause must string", c.Uuid)
		return nil
	}

	_, err := c.SndMsg("hangup", cause, true, true)
	if err != nil {
		logger.Error("Call %s hangup error: %v", c.Uuid, err.Error())
		return err
	}
	c.Conn.Disconnected = true
	logger.Debug("Call %s hangup cause: %v successful", c.Uuid, cause)
	return nil

}
