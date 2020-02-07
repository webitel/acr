/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

func RecordSession(scope Scope, c *Call, args interface{}) error {

	var action, typeFile, name, minSec, stereo, bridged, followTransfer string
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
		minSec = "2"
		stereo = "true"
		bridged = "true"
		followTransfer = "false"

	case map[string]interface{}:
		if prop, ok := args.(map[string]interface{}); ok {
			action = getStringValueFromMap("action", prop, "start")
			typeFile = getStringValueFromMap("type", prop, "mp3")
			name = getStringValueFromMap("name", prop, "recordSession")

			minSec = getStringValueFromMap("minSec", prop, "2")
			stereo = getStringValueFromMap("stereo", prop, "true")
			bridged = getStringValueFromMap("bridged", prop, "true")
			followTransfer = getStringValueFromMap("followTransfer", prop, "false")

		} else {
			//TODO
			return nil
		}
	}

	fileName := "${uuid}_" + name + "." + typeFile

	c.SetVariable("all:" + WEBITEL_RECORD_FILE_NAME + "=" + fileName)
	//c.Execute("export", "record_post_process_exec_app=lua:RecordUpload.lua ${uuid} ${domain_name} "+typeFile+" "+email+
	//	" "+name+" "+UrlEncoded(emailSubjectTemplate)+" "+UrlEncoded(emailTextTemplate))

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

		err = c.Execute("record_session", getRecordLink(c.DomainId(), c.Id(), fileName, typeFile))
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
