/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

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
		followTransfer = "false"

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
			followTransfer = getStringValueFromMap("followTransfer", prop, "false")

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

	SetVar(c, "all:"+WEBITEL_RECORD_FILE_NAME+"="+fileName)
	c.Execute("export", "record_post_process_exec_app=lua:RecordUpload.lua ${uuid} ${domain_name} "+typeFile+" "+email+
		" "+name+" "+UrlEncoded(emailSubjectTemplate)+" "+UrlEncoded(emailTextTemplate))

	if action == "start" {
		err = multiSetVar(c, []string{
			"RECORD_MIN_SEC=" + minSec,
			"RECORD_STEREO=" + stereo,
			"RECORD_BRIDGE_REQ=" + bridged,
			"recording_follow_transfer=" + followTransfer,
		})

		if err != nil {
			return err
		}

		err = c.Execute("record_session", "/recordings/"+fileName)
		if err != nil {
			c.LogError("recordSession", "/recordings/"+fileName, err.Error())
			return err
		}
		c.LogDebug("recordSession", "/recordings/"+fileName, "success")

	} else {

		err = c.Execute("stop_record_session", "/recordings/"+fileName)
		if err != nil {
			c.LogError("recordSession", "stop_record_session", err.Error())
			return err
		}
		c.LogDebug("recordSession", "stop_record_session", "success")
	}

	return nil
}
