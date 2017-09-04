/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func InBandDTMF(c *Call, args interface{}) error {
	var stopDTMF bool
	var app string
	if _, stopDTMF = args.(string); stopDTMF {
		if args.(string) == "stop" {
			stopDTMF = true
		} else {
			stopDTMF = false
		}
	}

	if stopDTMF {
		app = "stop_dtmf"
	} else {
		app = "start_dtmf"
	}

	_, err := c.SndMsg(app, "", true, true)
	if err != nil {
		logger.Error("Call %s inBandDTMF %s error: %s", c.Uuid, app, err.Error())
		return err
	}
	logger.Debug("Call %s inBandDTMF %v successful", c.Uuid, app)
	return nil
}
