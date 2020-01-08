/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

const WEBITEL_RECORD_FILE_NAME = "webitel_record_file_name"

func RecordFile(scope Scope, c *Call, args interface{}) error {

	var parameters map[string]interface{}
	var ok bool
	var err error

	if parameters, ok = args.(map[string]interface{}); !ok {
		c.LogError("recordFile", args, "bad request")
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

	if v := c.GetVariable(WEBITEL_RECORD_FILE_NAME); v != "" {
		err = c.Execute("stop_record_session", "/recordings/"+v)
		if err != nil {
			c.LogError("recordFile", "stop_record_session", err.Error())
			return err
		}
	}

	vars := []string{
		"playback_terminators=" + terminators,
		"record_post_process_exec_api=luarun:RecordFileUpload.lua ${uuid} ${domain_name} " + typeFile + " " + email + " " + name +
			" " + UrlEncoded(emailSubjectTemplate) + " " + UrlEncoded(emailTextTemplate),
	}
	err = multiSetVar(c, vars)

	if err != nil {
		c.LogError("recordFile", parameters, err.Error())
		return err
	}

	err = c.Execute("record", "/recordings/${uuid}_"+name+"."+typeFile+" "+maxSec+" "+silenceThresh+" "+silenceHits)
	if err != nil {
		c.LogError("recordFile", "/recordings/${uuid}_"+name+"."+typeFile+" "+maxSec+" "+silenceThresh+" "+silenceHits, err.Error())
		return err
	}

	return nil
}
