/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"strings"
)

//TODO ADD old goto support - delete new version;
func GoTo(c *Call, args interface{}) error {
	var tag string
	var ok bool

	if tag, ok = args.(string); !ok || tag == "" {
		logger.Error("Call %s: bad goto application arguments %s", c.Uuid, args)
		return nil
	}

	if strings.HasPrefix(tag, "local:") {
		if c.Iterator.Goto(tag[6:]) {
			logger.Debug("Call %s go to tag %s", c.Uuid, tag)
		} else {
			logger.Warning("Call %s not found tag %s", c.Uuid, tag)
		}

		return nil
	} else if strings.HasPrefix(tag, "default:") {
		tag = tag[8:] + " XML default"
	} else if strings.HasPrefix(tag, "public:") {
		tag = tag[7:] + " XML public"
	} else if strings.HasPrefix(tag, "socket:") {
		tag = tag[7:]
		_, err := c.SndMsg("socket", tag, false, false)
		if err != nil {
			logger.Error("Call %s socket error: ", c.Uuid, err)
			return err
		}
		c.SetBreak()
		c.Conn.Close()
		return nil

	} else {
		logger.Error("Call %s bad parameters: %s", c.Uuid, args)
		return nil
	}

	_, err := c.SndMsg("transfer", tag, false, false)
	if err != nil {
		logger.Error("Call %s transfer error: ", c.Uuid, err)
		return err
	}
	c.SetBreak()
	logger.Debug("Call %s transfer %s successful", c.Uuid, tag)

	return nil
}
