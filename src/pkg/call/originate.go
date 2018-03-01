package call

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
	"strings"
)

func Originate(c *Call, args interface{}) error {
	var props, endpoint, exportVar map[string]interface{}
	var ok bool
	var actions models.ArrayApplications
	var err error

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s originate bad arguments %v", c.Uuid, args)
		return nil
	}

	if _, ok = props["endpoint"]; !ok {
		logger.Error("Call %s originate endpoint is required", c.Uuid)
		return nil
	}

	if _, ok = props["actions"]; !ok {
		logger.Error("Call %s originate actions is required", c.Uuid)
		return nil
	}

	if _, ok = props["endpoint"].(map[string]interface{}); ok {
		endpoint = props["endpoint"].(map[string]interface{})
	} else {
		logger.Error("Call %s originate bad arguments endpoint %v", c.Uuid, props["endpoint"])
		return nil
	}

	if actions, ok = getArrayFromMap(props["actions"]); !ok {
		logger.Error("Call %s originate bad arguments actions %v", c.Uuid, props["actions"])
		return nil
	}

	uuidCall := getStringValueFromMap("uuid", props, "")
	if uuidCall == "" {
		uuidCall = uuid.New().String()
	} else {
		uuidCall = c.ParseString(uuidCall)
	}

	if uuidCall == "" {
		logger.Error("Call %s originate uuid is requred", c.Uuid)
		return nil
	}

	delay := getIntValueFromMap("delay", props, 2)
	timeout := getIntValueFromMap("timeout", props, 40)

	err = c.acr.InsertPrivateCallFlow(uuidCall, c.Domain, c.Timezone, timeout+delay+60, actions)
	if err != nil {
		logger.Error("Call %s originate create callflow db error: %s", c.Uuid, err.Error())
		return nil
	}

	var dialString = fmt.Sprintf("sched_api +%d none originate {origination_uuid=%s,originate_timeout=%d,domain_name=%s,webitel_direction=private,", delay,
		uuidCall, timeout, c.Domain)

	if _, ok = props["exportVar"]; ok {
		if exportVar, ok = props["exportVar"].(map[string]interface{}); ok {
			var k string
			for k, _ = range exportVar {
				dialString += k + "='" + c.ParseString(getStringValueFromMap(k, exportVar, "")) + "',"
			}
		}
	}

	dialString = dialString[:len(dialString)-1] + "}"

	addOriginateEndpoint(c, &dialString, endpoint)

	dialString += fmt.Sprintf(" %s XML private '%s' '%s'", uuidCall, c.ParseString(getStringValueFromMap("cid_num", props, "${caller_id_number}")),
		c.ParseString(getStringValueFromMap("cid_name", props, "${caller_id_name}")))

	dialString = c.ParseString(dialString)
	logger.Debug("Call %s originate: %s", c.Uuid, dialString)
	res, _ := c.Conn.Api(dialString)
	if strings.HasPrefix(string(res), "-ERR") {
		err = c.acr.RemovePrivateCallFlow(uuidCall, c.Domain)
		if err != nil {
			logger.Error("Call %s originate remove tmp callflow from db error: %s", c.Uuid, err.Error())
		}

		logger.Debug("Call %s originate error cause: %s", c.Uuid, res)
	} else {
		logger.Debug("Call %s originate to %s successful", c.Uuid, uuidCall)
	}

	return nil
}


func getProtoParameter(name string) string {
	return "[webitel_call_uuid=${create_uuid()},sip_invite_domain=${domain_name},presence_id=" + name + "@${domain_name}"
}
//TODO
func addOriginateEndpoint(c *Call, dialString *string, endpoint map[string]interface{}) {
	var ok bool
	var tmpArr []string

	switch getStringValueFromMap("type", endpoint, "") {
	case "sipGateway":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			*dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "sofia/gateway/" + getStringValueFromMap("name", endpoint, "_undef_") + "/" +
			getStringValueFromMap("dialString", endpoint, "_undef_")

	case "sipUri":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			*dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "external") + "/" +
			getStringValueFromMap("dialString", endpoint, "_undef_") + "@" + getStringValueFromMap("host", endpoint, "")

	case "sipDevice":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			*dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "external") + "/" +
			getStringValueFromMap("name", endpoint, "_undef_") + "%" +
			getStringValueFromMap("domainName", endpoint, "") + "^" + getStringValueFromMap("dialString", endpoint, "")

	case "device":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			*dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "user/" + getStringValueFromMap("name", endpoint, "_undef") + "@${domain_name}"

	case "user":
		switch getStringValueFromMap("proto", endpoint, "") {
		case "sip":
			*dialString += getProtoParameter(getStringValueFromMap("name", endpoint, "_undef_"))

			if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
				*dialString += "," + strings.Join(validateArrayVariables(tmpArr), ",")
			}
			*dialString += "]${sofia_contact(*/" + getStringValueFromMap("name", endpoint, "_undef") +
				"@${domain_name})}"

		case "webrtc":
			*dialString += getProtoParameter(getStringValueFromMap("name", endpoint, "_undef_"))

			if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
				*dialString += "," + strings.Join(validateArrayVariables(tmpArr), ",")
			}

			*dialString += "]${verto_contact(" + getStringValueFromMap("name", endpoint, "_undef_") +
				"@${domain_name})}"

		default:
			if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
				*dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
			}

			*dialString += "user/" + getStringValueFromMap("name", endpoint, "_undef_") + "@" +
				getStringValueFromMap("domainName", endpoint, "${domain_name}")
		}
	}
	return
}