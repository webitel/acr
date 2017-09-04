/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func FlushDTMF(c *Call, args interface{}) error {
	_, err := c.SndMsg("flush_dtmf", "", true, true)
	if err != nil {
		logger.Error("Call %s flushDTMF error: %s", c.Uuid, err.Error())
		return err
	}
	logger.Debug("Call %s flushDTMF successful", c.Uuid)
	return nil
}
