/**
 * Created by I. Navrotskyj on 25.09.17.
 */
package call

import (
	"encoding/json"
	"github.com/webitel/acr/src/pkg/logger"
	"strconv"
)

type communication struct {
	Number      string  `json:"number"`
	Priority    int     `json:"priority"`
	Status      int     `json:"status"`
	State       int     `json:"state"`
	Type        *string `json:"type"`
	Description string  `json:"description"`
}

type memberT struct {
	CreatedOn      int64                  `json:"createdOn" bson:"createdOn"`
	Name           string                 `json:"name"`
	Dialer         string                 `json:"dialer"`
	Domain         string                 `json:"domain"`
	Priority       int                    `json:"priority"`
	Expire         int                    `json:"expire"`
	Variables      map[string]interface{} `json:"variables"`
	Communications []communication        `json:"communications"`
}

func Member(c *Call, args interface{}) error {
	var tmp string
	var props map[string]interface{}
	var ok bool
	var err error

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s member bad arguments %s", c.Uuid, args)
		return nil
	}

	if _, ok = props["expire"]; ok {
		tmp = getStringValueFromMap("expire", props, "")
		if tmp != "" {
			props["expire"], _ = strconv.Atoi(c.ParseString(tmp))
		}
	}

	tmp = getStringValueFromMap("id", props, "")

	var b []byte
	b, err = json.Marshal(props)

	m := &memberT{}
	json.Unmarshal([]byte(c.ParseString(string(b))), m)
	m.CreatedOn = c.GetDate().Unix() * 1000
	if tmp != "" {
		delete(props, "id")
		err = c.acr.UpdateMember(tmp, &m)
		tmp = "update"
	} else {
		m.Domain = c.Domain
		if m.Dialer == "" {
			logger.Error("Call %s member argument dialer is required", c.Uuid, args)
			return nil
		}
		err = c.acr.AddMember(&m)
		tmp = "add"
	}

	if err != nil {
		logger.Error("Call %s: %s member: %s", c.Uuid, tmp, err.Error())
	} else {
		logger.Debug("Call %s: %s member successful", c.Uuid, tmp)
	}
	return nil
}
