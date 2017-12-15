/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"strings"
)

const WEBITEL_RECORD_FILE_NAME = "webitel_record_file_name"

func RecordFile(c *Call, args interface{}) error {

	var parameters map[string]interface{}
	var ok bool
	var err error

	if parameters, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s recordFile bad arguments: %v", c.Uuid, args)
		return nil
	}

	var name = getStringValueFromMap("name", parameters, "recordFile")
	var terminators = getStringValueFromMap("terminators", parameters, "#")
	var typeFile = getStringValueFromMap("type", parameters, "mp3")
	var maxSec = getStringValueFromMap("maxSec", parameters, "60")
	var silenceThresh = getStringValueFromMap("silenceThresh", parameters, "200")
	var silenceHits = getStringValueFromMap("silenceHits", parameters, "5")
	var email = "none"
	var emailTextTemplate, emailSubjectTemplate string

	if _, ok = parameters["email"]; ok {
		email = parseEmail(parameters["email"])
		emailTextTemplate = c.ParseString(getStringValueFromMap("emailBody", parameters, ""))
		emailSubjectTemplate = c.ParseString(getStringValueFromMap("emailSubject", parameters, ""))
	}

	if emailTextTemplate == "" {
		emailTextTemplate = "none"
	}
	if emailSubjectTemplate == "" {
		emailSubjectTemplate = "none"
	}

	if v := c.GetChannelVar(WEBITEL_RECORD_FILE_NAME); v != "" {
		_, err = c.SndMsg("stop_record_session", "/recordings/"+v, true, false)
		if err != nil {
			logger.Error("Call %s prev stop_record_session error: %s", c.Uuid, err.Error())
			return err
		}
	}

	err = SetVar(c, []interface{}{
		"playback_terminators=" + terminators,
		"record_post_process_exec_api=luarun:RecordUpload.lua ${uuid} ${domain_name} " + typeFile + " " + email + " " + name +
			" " + UrlEncoded(emailSubjectTemplate) + " " + UrlEncoded(emailTextTemplate),
	})

	if err != nil {
		logger.Error("Call %s set recordFile vars error: %s", c.Uuid, err.Error())
		return err
	}

	_, err = c.SndMsg("record", "/recordings/${uuid}_"+name+"."+typeFile+" "+maxSec+" "+silenceThresh+" "+silenceHits, true, false)
	if err != nil {
		logger.Error("Call %s record error: %s", c.Uuid, err.Error())
		return err
	}

	return nil
}

func parseEmail(parameters interface{}) string {
	var ok bool

	switch parameters.(type) {
	case string:
		return parameters.(string)

	case []interface{}:
		var email = ""
		for _, v := range parameters.([]interface{}) {
			if _, ok = v.(string); ok {
				email += "," + v.(string)
			}
		}
		if len(email) > 0 {
			email = email[1:]
		}

		if email == "" {
			email = "none"
		}
		return email

	case []string:
		return strings.Join(parameters.([]string), ",")
	}
	return "none"
}
