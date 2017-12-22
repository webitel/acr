/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
)

func RecordSession(c *Call, args interface{}) error {

	var action, typeFile, name, email, minSec, stereo, bridged, followTransfer, emailTextTemplate, emailSubjectTemplate string
	var err error

	switch args.(type) {
	case string:
		if args.(string) == "stop" {
			action = "stop"
		} else {
			action = "start"
		}

		typeFile = "mp3"
		name = "recordSession"
		email = "none"
		minSec = "2"
		stereo = "true"
		bridged = "true"
		followTransfer = "true"

	case map[string]interface{}:
		if prop, ok := args.(map[string]interface{}); ok {
			action = getStringValueFromMap("action", prop, "start")
			typeFile = getStringValueFromMap("type", prop, "mp3")
			name = getStringValueFromMap("name", prop, "recordSession")

			if _, ok = prop["email"]; ok {
				email = parseEmail(prop["email"])
				emailTextTemplate = c.ParseString(getStringValueFromMap("emailBody", prop, ""))
				emailSubjectTemplate = c.ParseString(getStringValueFromMap("emailSubject", prop, ""))
			} else {
				email = "none"
			}

			minSec = getStringValueFromMap("minSec", prop, "2")
			stereo = getStringValueFromMap("stereo", prop, "true")
			bridged = getStringValueFromMap("bridged", prop, "true")
			followTransfer = getStringValueFromMap("followTransfer", prop, "true")

		} else {
			//TODO
			return nil
		}
	}

	if emailTextTemplate == "" {
		emailTextTemplate = "none"
	}
	if emailSubjectTemplate == "" {
		emailSubjectTemplate = "none"
	}

	fileName := "${uuid}_" + name + "." + typeFile

	if action == "start" {
		err = multiSetVar(c, []string{
			WEBITEL_RECORD_FILE_NAME + "=" + fileName,
			"RECORD_MIN_SEC=" + minSec,
			"RECORD_STEREO=" + stereo,
			"RECORD_BRIDGE_REQ=" + bridged,
			"recording_follow_transfer=" + followTransfer,
			"record_post_process_exec_api=luarun:RecordUpload.lua ${uuid} ${domain_name} " + typeFile + " " + email +
				" " + name + " " + UrlEncoded(emailSubjectTemplate) + " " + UrlEncoded(emailTextTemplate),
		})

		if err != nil {
			logger.Error("Call %s set recordSession vars error: %s", c.Uuid, err.Error())
			return err
		}

		_, err = c.SndMsg("record_session", "/recordings/"+fileName, true, false)
		if err != nil {
			logger.Error("Call %s start record_session error: %s", c.Uuid, err.Error())
			return err
		}
		logger.Debug("Call %s start record_session %s successful", c.Uuid, fileName)

	} else {

		_, err = c.SndMsg("stop_record_session", "/recordings/"+fileName, true, false)
		if err != nil {
			logger.Error("Call %s stop_record_session error: %s", c.Uuid, err.Error())
			return err
		}
		logger.Debug("Call %s stop_record_session %s successful", c.Uuid, fileName)
	}

	return nil
}
