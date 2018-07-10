/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"crypto/tls"
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/gomail.v2"
)

type emailConfig struct {
	Provider string `bson:"provider"`
	From     string `bson:"from"`

	Host     string `bson:"host"`
	User     string `bson:"user"`
	Password string `bson:"pass"`
	Secure   bool   `bson:"secure"`
	Port     int    `bson:"port"`
}

func SendEmail(c *Call, args interface{}) error {
	var conf emailConfig
	var props map[string]interface{}
	var ok bool

	var message, tmp string
	var to []string

	if props, ok = args.(map[string]interface{}); ok {
		if message = getStringValueFromMap("message", props, ""); message == "" {
			logger.Error("Call %s sendEmail message is required", c.Uuid)
			return nil
		}

		if _, ok = props["to"]; !ok {
			logger.Error("Call %s sendEmail to is required", c.Uuid)
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
			logger.Error("Call %s sendEmail bad to: %v", c.Uuid, props["to"])
			return nil
		}

		if err := c.acr.GetEmailConfig(c.Domain, &conf); err != nil {
			logger.Error("Call %s sendEmail db error: %s", err.Error())
			return nil
		}

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
				logger.Error("Call %s sendMail error: %s", uuid, err.Error())
			} else {
				logger.Debug("Call %s sendMail to: %v successful", uuid, to)
			}
		}(c.Uuid, d, m, to)
	} else {
		logger.Error("Call %s sendEmail bad arguments %s", c.Uuid, args)
	}

	return nil
}
