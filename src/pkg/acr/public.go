/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package acr

import (
	"github.com/webitel/acr/src/pkg/call"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
)

var defaultPublicRoute = config.Conf.Get("defaultPublicRout")

func publicContext(a *ACR, c *esl.SConn, destinationNumber string) {

	var cf models.CallFlow
	var def string
	var err error

	cf, err = a.DB.FindPublic(destinationNumber)
	if err != nil {
		logger.Error("Call %s db error: %s", c.Uuid, err.Error())
		c.Hangup(HANGUP_NORMAL_TEMPORARY_FAILURE)
		return
	}

	if cf.Id != 0 {
		createPublicCall(a, c, destinationNumber, &cf)
		return
	}

	if defaultPublicRoute != "" && defaultPublicRoute != "<nil>" {
		def = defaultPublicRoute
	} else {
		def, _ = a.GetGlobalVarBySwitchId(c.ChannelData.Header.Get("Core-UUID"), "webitel_default_public_route")
	}

	if def != "" {
		cf, err = a.DB.FindPublic(def)
		if err != nil {
			logger.Error("Call %s db error: %s", c.Uuid, err.Error())
			c.Hangup(HANGUP_NORMAL_TEMPORARY_FAILURE)
			return
		}
		if cf.Id != 0 {
			createPublicCall(a, c, def, &cf)
			return
		}
	}

	if setDirection(c, "inbound") != nil {
		return
	}

	logger.Warning("Call %s: no found public context number %s", c.Uuid, destinationNumber)
	c.Hangup(HANGUP_NO_ROUTE_DESTINATION)
}

func createPublicCall(a *ACR, c *esl.SConn, destinationNumber string, cf *models.CallFlow) {
	var err error

	if setDirection(c, "inbound") != nil {
		return
	}

	if cf.Timezone != "" {
		_, err = c.SndMsg("set", "timezone="+cf.Timezone, false, false)
		if err != nil {
			logger.Error("Call %s error: %s", c.Uuid, err.Error())
			return
		}
		logger.Debug("Call %s set timezone=%s", c.Uuid, cf.Timezone)
	}

	_, err = c.SndMsg("set", "domain_name="+cf.Domain, false, true)
	if err != nil {
		logger.Error("Call %s error: %s", c.Uuid, err.Error())
		return
	}
	_, err = c.SndMsg("set", "force_transfer_context=default", false, false)
	if err != nil {
		logger.Error("Call %s error: %s", c.Uuid, err.Error())
		return
	}

	a.CreateCall(destinationNumber, c, cf, call.CONTEXT_PUBLIC)
	return
}
