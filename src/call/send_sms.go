/**
 * Created by I. Navrotskyj on 30.08.17.
 */

package call

import (
	"bytes"
	"fmt"
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
		c.LogError("sendSMS", args, "bad request")
		return nil
	}

	if _, ok = props["phone"]; !ok {
		c.LogError("sendSMS", props, "phone is require")
		return nil
	}

	if _, ok = props["phone"].(string); ok {
		phones = []string{props["phone"].(string)}
	} else if phones, ok = getArrayStringFromMap("phone", props); !ok {
		c.LogError("sendSMS", props, "bad format phone")
		return nil
	}

	countPhones = len(phones)

	if countPhones == 0 {
		c.LogError("sendSMS", props, "bad format phone")
		return nil
	}

	if login = getStringValueFromMap("login", props, ""); login == "" {
		c.LogError("sendSMS", props, "login is require")
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
		c.LogError("sendSMS", xml, err.Error())
		return SetVar(c, "sendSms=false")
	}

	req.Header.Set("Content-Type", "application/xml")

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	res, err = client.Do(req)
	if err != nil {
		c.LogError("sendSMS", xml, err.Error())
		return SetVar(c, "sendSms=false")
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		c.LogDebug("sendSMS", xml, "success")
		return SetVar(c, "sendSms=true")
	} else {
		c.LogError("sendSMS", xml, fmt.Sprintf("response code: %v", res.StatusCode))
		return SetVar(c, "sendSms=false")
	}
}
