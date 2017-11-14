/**
 * Created by I. Navrotskyj on 08.11.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
	"strings"
)

func SendEvent(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s bad arguments %v", c.Uuid, args)
		return nil
	}
	var m esl.Message
	m = esl.Message{
		Header: esl.Header{
			"Event-Subclass": []string{"ACR::HOOK"},
		},
	}

	if getBoolValueFromMap("dump", props, true) {
		for k, v := range c.Conn.ChannelData.Header {
			if isProtectedHeader(k) {
				continue
			}

			if len(v) > 0 {
				m.Header.Add(k, strings.Replace(v[0], "\n", " ", -1))
			}
		}
	}

	if _, ok = props["data"]; ok {
		if _, ok = props["data"].(map[string]interface{}); ok {
			data := props["data"].(map[string]interface{})
			for k, _ := range data {
				m.Header.Add(k, c.ParseString(getStringValueFromMap(k, data, "")))
			}
		}
	}
	m.Header.Del("Event-Name")
	m.Header.Add("variable_domain_name", c.Domain)

	_, err := c.Conn.FireEvent("custom", &m)
	if err != nil {
		logger.Error("Call %s sendEvent error: %s", c.Uuid, err.Error())
		return err
	}
	logger.Debug("Call %s sendEvent successful", c.Uuid)
	return nil
}

func isProtectedHeader(name string) bool {
	switch name {
	case "Event-Subclass", "variable_domain_name", "Event-Name", "Content-Type", "Reply-Text", "Presence-Call-Direction", "Core-UUID":
		return true
	}

	return false
}
