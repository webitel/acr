/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package acr

import (
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
)

const PUBLIC_CONTEXT = "public"
const DEFAULT_CONTEXT = "default"
const DIALER_CONTEXT = "dialer"
const PRIVATE_CONTEXT = "private"

func (a *ACR) routeContext(c *esl.SConn) {
	setSoundLang(c)
	//TODO move destination number to socket
	dialerId := c.ChannelData.Header.Get("variable_dlr_queue")
	destinationNumber := c.ChannelData.Header.Get("Channel-Destination-Number")
	if destinationNumber == "" {
		destinationNumber = c.ChannelData.Header.Get("Caller-Destination-Number")
		if destinationNumber == "" {
			destinationNumber = c.ChannelData.Header.Get("variable_destination_number")
		}
	}

	switch c.GetContextName() {
	case PUBLIC_CONTEXT:
		logger.Debug("Call %s from context public to %s", c.Uuid, destinationNumber)
		publicContext(a, c, destinationNumber)
	case DIALER_CONTEXT:
		logger.Debug("Call %s from context dialer (%s) to %s", c.Uuid, dialerId, destinationNumber)
		dialerContext(a, c, destinationNumber, dialerId)
	case DEFAULT_CONTEXT:
		logger.Debug("Call %s from context default to %s", c.Uuid, destinationNumber)
		defaultContext(a, c, destinationNumber)
	case PRIVATE_CONTEXT:
		logger.Debug("Call %s from context private to %s", c.Uuid, destinationNumber)
		privateContext(a, c, destinationNumber)
	default:
		logger.Debug("Call %s: no found context %s", c.Uuid, c.GetContextName())
		c.Hangup(HANGUP_NO_ROUTE_DESTINATION)
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
