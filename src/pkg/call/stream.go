/**
 * Created by I. Navrotskyj on 24.10.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func Stream(c *Call, args interface{}) error {
	var uri = ""
	switch args.(type) {
	case string:
		uri = args.(string)
	default:
		logger.Error("Call %s stream cause must string", c.Uuid)
		return nil
	}

	//_, err := c.SndMsg("record_session", uri, true, true)
	//_, err := c.SndMsg("record", uri, true, true)
	_, err := c.SndMsg("playback", uri, true, true)
	if err != nil {
		logger.Error("Call %s cause error: %v", c.Uuid, err.Error())
		return err
	}

	return nil
}
