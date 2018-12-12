/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package acr

import (
	"github.com/webitel/acr/src/pkg/call"
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
)

func defaultContext(a *ACR, c *esl.SConn, destinationNumber string) {
	domainName := c.ChannelData.Header.Get("variable_domain_name")

	_, err := c.SndMsg("unset", "sip_h_call-info", false, true)
	if err != nil {
		logger.Error("Call %s bad unset sip_h_call-info: %s", c.Uuid, err.Error())
	}

	cf := models.CallFlow{}

	cf, err = a.DB.FindExtension(destinationNumber, domainName)
	if err != nil {
		logger.Error("Call %s db error: %s", c.Uuid, err.Error())
		c.Hangup(HANGUP_NORMAL_TEMPORARY_FAILURE)
		return
	}

	if cf.Id != 0 {
		internalCall(destinationNumber, a, c, &cf)
		return
	}

	cf, err = a.DB.FindDefault(domainName, destinationNumber)
	if err != nil {
		logger.Error("Call %s db error: %s", c.Uuid, err.Error())
		c.Hangup(HANGUP_NORMAL_TEMPORARY_FAILURE)
		return
	}

	if cf.Id != 0 {
		worldCall(destinationNumber, a, c, &cf)
		return
	}

	if setDirection(c, "outbound") != nil {
		return
	}

	logger.Debug("Call %s: no found default context number %s", c.Uuid, destinationNumber)
	c.Hangup(HANGUP_NO_ROUTE_DESTINATION)
}

func internalCall(destinationNumber string, a *ACR, c *esl.SConn, cf *models.CallFlow) {
	logger.Debug("Call %s is internal", c.Uuid)
	var err error

	if setDirection(c, "internal") != nil {
		return
	}

	_, err = c.SndMsg("export", "nolocal:sip_redirect_context=default", false, false)
	if err != nil {
		logger.Error("Call %s bad export sip_redirect_context: %s", c.Uuid, err.Error())
	}

	if cf.Timezone != "" {
		_, err = c.SndMsg("set", "timezone="+cf.Timezone, false, false)
		if err != nil {
			logger.Error("Call %s bad set timezone: %s", c.Uuid, err.Error())
		} else {
			logger.Debug("Call %s set timezone %s", c.Uuid, cf.Timezone)
		}
	}

	a.CreateCall(destinationNumber, c, cf, call.CONTEXT_DEFAULT)
}

func worldCall(destinationNumber string, a *ACR, c *esl.SConn, cf *models.CallFlow) {
	logger.Debug("Call %s is default context %s %s", c.Uuid, cf.Name, cf.Number)

	if setDirection(c, "outbound") != nil {
		return
	}

	a.CreateCall(destinationNumber, c, cf, call.CONTEXT_DEFAULT)
}
