/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"crypto/tls"
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/wlog"
	"gopkg.in/gomail.v2"
)

func SendEmail(c *Call, args interface{}) error {
	var conf *model.EmailConfig
	var props map[string]interface{}
	var ok bool

	var message, tmp string
	var to []string

	if props, ok = args.(map[string]interface{}); ok {
		if message = getStringValueFromMap("message", props, ""); message == "" {
			c.LogError("email", props, "message is require")
			return nil
		}

		if _, ok = props["to"]; !ok {
			c.LogError("email", props, "to is require")
			return nil
		} else {
			switch props["to"].(type) {
			case string:
				to = []string{c.ParseString(props["to"].(string))}
			case []interface{}:
				for _, v := range props["to"].([]interface{}) {
					if tmp, ok = v.(string); ok && tmp != "" {
						to = append(to, c.ParseString(tmp))
					}
				}
			}
		}

		if len(to) == 0 {
			c.LogError("email", props, "bad 'to'")
			return nil
		}

		result := <-c.router.app.Store.Email().Config(c.Domain())
		if result.Err != nil {
			c.LogError("email", props, result.Err.Error())
			return nil
		}

		conf = result.Data.(*model.EmailConfig)

		message = c.ParseString(message)

		m := gomail.NewMessage()
		if tmp = getStringValueFromMap("from", props, conf.From); tmp != "" {
			m.SetHeader("From", tmp)
		}

		m.SetHeader("To", to...)

		if tmp = getStringValueFromMap("subject", props, ""); tmp != "" {
			m.SetHeader("Subject", c.ParseString(tmp))
		}

		m.SetBody("text/html", message)

		d := gomail.NewDialer(conf.Host, conf.Port, conf.User, conf.Password)
		if !conf.Secure {
			d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		}

		go func(uuid string, d *gomail.Dialer, m *gomail.Message, to []string) {
			// Send the email to Bob, Cora and Dan.
			if err := d.DialAndSend(m); err != nil {
				wlog.Error(fmt.Sprintf("call %s send email error %s", uuid, err.Error()))
			} else {
				wlog.Debug(fmt.Sprintf("call %s send email to %v success", uuid, to))
			}
		}(c.Id(), d, m, to)
	} else {
		c.LogError("email", args, "bad request")
	}

	return nil
}
