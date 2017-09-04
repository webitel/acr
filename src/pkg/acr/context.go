/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package acr

import (
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
)

const PUBLIC_CONTEXT = "public"

func (a *ACR) routeContext(c *esl.SConn) {
	setSoundLang(c)
	context := c.ChannelData.Header.Get("Channel-Context")
	dialerId := c.ChannelData.Header.Get("variable_dlr_queue")
	destinationNumber := c.ChannelData.Header.Get("Channel-Destination-Number")
	if destinationNumber == "" {
		destinationNumber = c.ChannelData.Header.Get("Caller-Destination-Number")
		if destinationNumber == "" {
			destinationNumber = c.ChannelData.Header.Get("variable_destination_number")
		}
	}

	if context == PUBLIC_CONTEXT {
		logger.Debug("Call %s from context public to %s", c.Uuid, destinationNumber)
		publicContext(a, c, destinationNumber)
	} else if dialerId != "" {
		logger.Debug("Call %s from context dialer (%s) to %s", c.Uuid, dialerId, destinationNumber)
		dialerContext(a, c, destinationNumber, dialerId)
	} else {
		logger.Debug("Call %s from context default to %s", c.Uuid, destinationNumber)
		defaultContext(a, c, destinationNumber)
	}
}

func setSoundLang(c *esl.SConn) {
	var e error
	if c.ChannelData.Header.Get("variable_default_language") == "ru" {
		_, e = c.SndMsg("set", "sound_prefix=/$${sounds_dir}/ru/RU/elena", true, false)
	} else {
		_, e = c.SndMsg("set", "sound_prefix=/$${sounds_dir}/en/us/callie", true, false)
	}

	if e != nil {
		logger.Error("Set sound lang: ", e)
	}
}
