/**
 * Created by I. Navrotskyj on 30.08.17.
 */

package call

import (
	"bytes"
	"github.com/webitel/acr/src/pkg/logger"
	"net/http"
	"time"
)

func SendSMS(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var login, xml, tmp string
	var phones []string
	var countPhones int
	var req *http.Request
	var res *http.Response
	var err error

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s sendSms bad arguments %v", c.Uuid, args)
		return nil
	}

	if _, ok = props["phone"]; !ok {
		logger.Error("Call %s sendSms phone is required", c.Uuid)
		return nil
	}

	if _, ok = props["phone"].(string); ok {
		phones = []string{props["phone"].(string)}
	} else if phones, ok = getArrayStringFromMap("phone", props); !ok {
		logger.Error("Call %s sendSms bad format phone %v", c.Uuid, props["phone"])
		return nil
	}

	countPhones = len(phones)

	if countPhones == 0 {
		logger.Error("Call %s sendSms bad format phone %v", c.Uuid, props["phone"])
		return nil
	}

	if login = getStringValueFromMap("login", props, ""); login == "" {
		logger.Error("Call %s sendSms login is required", c.Uuid)
		return nil
	}

	xml = `<?xml version="1.0" encoding="UTF-8" ?>` + "\n" +
		`<request method="send-sms" login="` + login + `" passw="` + getStringValueFromMap("password", props, "") + `">` +
		"\n\t" + `<msg id="` + getStringValueFromMap("id", props, "1") + `" `

	if countPhones == 1 {
		xml += `phone="` + phones[0] + `" `
	}

	xml += `sn="` + getStringValueFromMap("name", props, "") + `" `

	if tmp = getStringValueFromMap("send_time", props, ""); tmp != "" {
		xml += `send_time="` + tmp + `" `
	}

	if tmp = getStringValueFromMap("encoding", props, ""); tmp != "" {
		xml += `encoding="` + tmp + `" `
	}

	xml += `>` + getStringValueFromMap("message", props, "") + `</msg>` + "\n"

	if countPhones > 1 {
		for _, tmp = range phones {
			xml += `<phone number="` + tmp + `"` + " />\n"
		}
	}

	xml += `</request>`
	xml = c.ParseString(xml)

	req, err = http.NewRequest("POST", "http://sms.barex.com.ua/websend/", bytes.NewBuffer([]byte(xml)))
	if err != nil {
		logger.Error("Call %s sendSms create request error: %s", c.Uuid, err.Error())
		return SetVar(c, "sendSms=false")
	}

	req.Header.Set("Content-Type", "application/xml")

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	res, err = client.Do(req)
	if err != nil {
		logger.Error("Call %s sendSms response error: %s", c.Uuid, err.Error())
		return SetVar(c, "sendSms=false")
	}
	defer res.Body.Close()
	logger.Debug("Call %s sendSms response code %v", c.Uuid, res.StatusCode)

	if res.StatusCode == 200 {
		logger.Debug("Call %s sendSms to %v successful", c.Uuid, phones)
		return SetVar(c, "sendSms=true")
	} else {
		logger.Debug("Call %s sendSms to %v error status code %v", c.Uuid, phones, res.StatusCode)
		return SetVar(c, "sendSms=false")
	}
}
