package acr

import (
	"github.com/webitel/acr/src/pkg/call"
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
)

func privateContext(a *ACR, c *esl.Connection, uuid string) {
	callFlow, err := a.DB.GetPrivateCallFlow(uuid, c.ChannelData.Header.Get("variable_domain_name"))
	if err != nil {
		logger.Error("Call %s find callflow db error: %s", c.Uuid, err.Error())
		c.Hangup(HANGUP_NORMAL_TEMPORARY_FAILURE)
		return
	}

	if len(callFlow.Callflow) == 0 {
		logger.Debug("Call %s: no found private context", c.Uuid)
		c.Hangup(HANGUP_NO_ROUTE_DESTINATION)
	}
	a.CreateCall(uuid, c, &callFlow, call.CONTEXT_PRIVATE)
}
