package acr

import (
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
)

func setDirection(c *esl.SConn, direction string) (err error) {
	if c.ChannelData.Header.Get("variable_webitel_direction") == "" {
		_, err = c.SndMsg("set", "webitel_direction="+direction, false, false)
		if err != nil {
			logger.Error("Call %s error: %s", c.Uuid, err.Error())
		} else {
			logger.Debug("Call %s set webitel_direction="+direction, c.Uuid)
		}
	}
	return
}
