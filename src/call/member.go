/**
 * Created by I. Navrotskyj on 25.09.17.
 */
package call

import (
	"encoding/json"
	"github.com/webitel/acr/src/model"
	"strconv"
)

func Member(c *Call, args interface{}) error {
	var tmp string
	var props map[string]interface{}
	var ok bool

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("member", args, "bad request")
		return nil
	}

	if _, ok = props["expire"]; ok {
		tmp = getStringValueFromMap("expire", props, "")
		if tmp != "" {
			props["expire"], _ = strconv.Atoi(c.ParseString(tmp))
		}
	}

	tmp = getStringValueFromMap("id", props, "")

	b, err := json.Marshal(props)
	if err != nil {
		c.LogError("member", props, err.Error())
		return nil
	}

	m := &model.OutboundQueueMember{}
	json.Unmarshal([]byte(c.ParseString(string(b))), m)
	m.CreatedOn = c.GetDate().Unix() * 1000
	if tmp != "" {
		delete(props, "id")
		result := <-c.router.app.Store.OutboundQueue().UpdateMember(tmp, m)
		if result.Err != nil {
			c.LogError("member", m, result.Err.Error())
			return nil
		}
		tmp = "update"
	} else {
		m.Domain = c.Domain()
		if m.Dialer == "" {
			c.LogError("member", m, "dialer is required")
			return nil
		}

		if nextCallSec := getIntValueFromMap("callAfterSec", props, 0); nextCallSec > 0 {
			var nextCall = model.GetMillis() + int64(nextCallSec*1000)
			m.NextCallAfterSec = &nextCall
		}

		result := <-c.router.app.Store.OutboundQueue().CreateMember(m)
		if result.Err != nil {
			c.LogError("member", m, result.Err.Error())
			return nil
		}
		tmp = "add"
	}

	c.LogDebug("member", props, "success")
	return nil
}
