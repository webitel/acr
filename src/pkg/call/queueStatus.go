package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"fmt"
)

func QueueStatus(c *Call, args interface{}) error {
	var props map[string]interface{}
	var queueName, tmp string
	var ok bool

	if props, ok = args.(map[string]interface{}); ok {
		queueName = getStringValueFromMap("name", props, "")
		if queueName == "" {
			logger.Error("Call %s queueStatus arguments name is required", c.Uuid)
			return nil
		}

		if tmp = getStringValueFromMap("freeAgents", props, ""); tmp != "" {
			if err := SetVar(c, fmt.Sprintf("%s=%d", tmp, c.acr.CountAvailableAgent(queueName + "@" + c.Domain))); err != nil {
				return err
			}
		}

		if tmp = getStringValueFromMap("waitingMembers", props, ""); tmp != "" {
			if err := SetVar(c, fmt.Sprintf("%s=%d", tmp, c.acr.CountAvailableMembers(queueName + "@" + c.Domain))); err != nil {
				return err
			}
		}

	} else {
		logger.Error("Call %s queueStatus bad arguments %v", c.Uuid, args)
	}
	return nil
}