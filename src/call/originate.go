package call

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/webitel/acr/src/model"
	"strings"
)

func Originate(scope Scope, c *Call, args interface{}) error {
	var props, endpoint, exportVar map[string]interface{}
	var ok bool
	var actions model.ArrayApplications
	var err error

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("originate", args, "bad request")
		return nil
	}

	if _, ok = props["endpoint"]; !ok {
		c.LogError("originate", props, "endpoint is required")
		return nil
	}

	if _, ok = props["actions"]; !ok {
		c.LogError("originate", props, "actions is required")
		return nil
	}

	if _, ok = props["endpoint"].(map[string]interface{}); ok {
		endpoint = props["endpoint"].(map[string]interface{})
	} else {
		c.LogError("originate", props, "bad arguments endpoint")
		return nil
	}

	if actions, ok = getArrayFromMap(props["actions"]); !ok {
		c.LogError("originate", props, "bad arguments actions")
		return nil
	}

	uuidCall := getStringValueFromMap("uuid", props, "")
	if uuidCall == "" {
		uuidCall = uuid.New().String()
	} else {
		uuidCall = c.ParseString(uuidCall)
	}

	if uuidCall == "" {
		c.LogError("originate", props, "uuid is required")
		return nil
	}

	delay := getIntValueFromMap("delay", props, 2)
	timeout := getIntValueFromMap("timeout", props, 40)

	result := <-c.router.app.Store.PrivateRoute().Create(uuidCall, c.Domain(), c.Timezone(), timeout+delay+60, actions)
	if result.Err != nil {
		c.LogError("originate", props, result.Err.Error())
		return nil
	}

	var dialString = fmt.Sprintf("sched_api +%d none originate {origination_uuid=%s,originate_timeout=%d,domain_name=%s,webitel_direction=private,", delay,
		uuidCall, timeout, c.Domain())

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

	res, err := c.Api(dialString)
	if err != nil {
		c.LogError("originate", dialString, err.Error())
		return nil
	}

	if strings.HasPrefix(string(res), "-ERR") {
		<-c.router.app.Store.PrivateRoute().Remove(c.Domain(), uuidCall)
		c.LogError("originate", dialString, string(res))
	} else {
		c.LogDebug("originate", dialString, "success")
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
			*dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "sofia/gateway/" + getStringValueFromMap("name", endpoint, "_undef_") + "/" +
			getStringValueFromMap("dialString", endpoint, "_undef_")

	case "sipUri":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			*dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "external") + "/" +
			getStringValueFromMap("dialString", endpoint, "_undef_") + "@" + getStringValueFromMap("host", endpoint, "")

	case "sipDevice":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			*dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "external") + "/" +
			getStringValueFromMap("name", endpoint, "_undef_") + "%" +
			getStringValueFromMap("domainName", endpoint, "") + "^" + getStringValueFromMap("dialString", endpoint, "")

	case "device":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			*dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "user/" + getStringValueFromMap("name", endpoint, "_undef") + "@${domain_name}"

	case "user":
		switch getStringValueFromMap("proto", endpoint, "") {
		case "sip":
			*dialString += getProtoParameter(getStringValueFromMap("name", endpoint, "_undef_"))

			if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
				*dialString += "," + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",")
			}
			*dialString += "]${sofia_contact(*/" + getStringValueFromMap("name", endpoint, "_undef") +
				"@${domain_name})}"

		case "webrtc":
			*dialString += getProtoParameter(getStringValueFromMap("name", endpoint, "_undef_"))

			if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
				*dialString += "," + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",")
			}

			*dialString += "]${verto_contact(" + getStringValueFromMap("name", endpoint, "_undef_") +
				"@${domain_name})}"

		default:
			if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
				*dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
			}

			*dialString += "user/" + getStringValueFromMap("name", endpoint, "_undef_") + "@" +
				getStringValueFromMap("domainName", endpoint, "${domain_name}")
		}
	}
	return
}
