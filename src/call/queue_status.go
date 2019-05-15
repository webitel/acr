package call

import (
	"fmt"
)

func QueueStatus(c *Call, args interface{}) error {
	var props map[string]interface{}
	var queueName, tmp string
	var ok bool

	if props, ok = args.(map[string]interface{}); ok {
		queueName = getStringValueFromMap("name", props, "")
		if queueName == "" {
			c.LogError("queueStatus", props, "name is required")
			return nil
		}

		if tmp = getStringValueFromMap("freeAgents", props, ""); tmp != "" {
			result := <-c.router.app.Store.InboundQueue().CountAvailableAgent(c.Domain(), queueName)
			if result.Err != nil {
				c.LogError("queueStatus", props, result.Err.Error())
			} else {
				err := SetVar(c, fmt.Sprintf("%s=%d", tmp, result.Data.(int)))
				if err != nil {
					return err
				}
			}
		}

		if tmp = getStringValueFromMap("waitingMembers", props, ""); tmp != "" {
			result := <-c.router.app.Store.InboundQueue().CountAvailableMembers(c.Domain(), queueName)
			if result.Err != nil {
				c.LogError("queueStatus", props, result.Err.Error())
			} else {
				err := SetVar(c, fmt.Sprintf("%s=%d", tmp, result.Data.(int)))
				if err != nil {
					return err
				}
			}
		}

	} else {
		c.LogError("queueStatus", args, "bad request")
	}
	return nil
}
