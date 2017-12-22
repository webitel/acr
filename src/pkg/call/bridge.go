/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
	"strings"
)

func Bridge(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s bridge bad arguments %s", c.Uuid, args)
		return nil
	}

	if _, ok = props["uuid"]; ok {
		_, ok = props["other_uuid"]
	}

	if ok {
		return bridgeUuid(c,
			getStringValueFromMap("uuid", props, ""),
			getStringValueFromMap("other_uuid", props, ""),
			getStringValueFromMap("exportVar", props, "uuid_bridge_res"),
		)
	} else {
		return bridgeChannel(c, props)
	}
}

func bridgeUuid(c *Call, legA, legB string, resVar string) error {

	if c.Conn.GetDisconnected() {
		logger.Debug("Call %s uuid bridge error: disconnected")
		return nil
	}

	legA = c.ParseString(legA)
	legB = c.ParseString(legB)

	if legA == "" || legB == "" {
		logger.Error("Call %s uuid bridge error: empty uuid or other_uuid", c.Uuid)
		return nil
	}

	bridgeString := fmt.Sprintf("uuid_bridge %s %s", legA, legB)
	logger.Debug("Call %s uuid bridge: %s", c.Uuid, bridgeString)

	resBytes, err := c.Conn.Api(bridgeString)
	if err != nil {
		logger.Error("Call %s uuid bridge error: %s", c.Uuid, err.Error())
		return nil
	}
	resStr := strings.Replace(string(resBytes), "\n", "", -1)

	logger.Debug("Call %s uuid bridge successful %s", c.Uuid, string(resStr))

	if c.Conn.GetDisconnected() {
		return nil
	}
	return SetVar(c, resVar+"="+resStr)
}

func bridgeChannel(c *Call, props map[string]interface{}) error {
	var ok bool
	var strategy, separator, dialString, pickup, p string
	var tmpArr, params []string
	var endpoints models.ArrayApplications

	if _, ok = props["endpoints"]; !ok {
		logger.Error("Call %s bridge endpoints is required", c.Uuid)
		return nil
	}

	if endpoints, ok = getArrayFromMap(props["endpoints"]); !ok {
		logger.Error("Call %s bridge bad endpoints parameter", c.Uuid)
		return nil
	}

	if len(endpoints) == 0 {
		logger.Error("Call %s bridge empty endpoints", c.Uuid)
		return nil
	}

	strategy = getStringValueFromMap("strategy", props, "")

	if strategy == "failover" {
		separator = "|"
	} else if strategy != "" && strategy != "multiple" {
		separator = strategy
	} else {
		separator = ":_:"
	}

	if tmpArr, ok = getArrayStringFromMap("global", props); ok && len(tmpArr) > 0 {
		dialString += "<" + strings.Join(validateArrayVariables(tmpArr), ",") + ">"
	}

	dialString += "{" + "domain_name=" + c.Domain

	if tmpArr, ok = getArrayStringFromMap("parameters", props); ok && len(tmpArr) > 0 {
		params = validateArrayVariables(tmpArr)
	}

	if tmpArr, ok = getArrayStringFromMap("codecs", props); ok && len(tmpArr) > 0 {
		params = append(params, "absolute_codec_string='"+strings.Join(tmpArr, ",")+"'")
	}

	if len(params) > 0 {
		dialString += "," + strings.Join(params, ",")
	}

	dialString += "}"

	for _, endpoint := range endpoints {
		addBridgeEndpoint(c, &dialString, endpoint)
		dialString += separator
	}

	if _, ok = props["pickup"]; strategy != "failover" && ok {
		p = getStringValueFromMap("pickup", props, "")
		if p == "" {
			if tmpArr, ok = getArrayStringFromMap("pickup", props); ok {
				for _, p = range tmpArr {
					pickup += ",pickup/" + p + "@${domain_name}"
				}
			}
		} else {
			pickup = ",pickup/" + p + "@${domain_name}"
		}
	}

	e, err := c.SndMsg("bridge", dialString[:len(dialString)-len(separator)]+pickup, true, true)
	if err != nil {
		logger.Error("Call %s bridge error: %s", c.Uuid, err)
		return err
	}

	if e.Header.Get("variable_bridge_hangup_cause") == "NORMAL_CLEARING" && e.Header.Get("variable_hangup_after_bridge") == "true" {
		c.SetBreak()
	}

	return nil
}

func setSpyMap(c *Call, name string) {
	c.SndMsg("hash", "insert/spymap/${domain_name}-"+name+"/${uuid}", true, false)
}

func getProtoParameter(name string) string {
	return "[webitel_call_uuid=${create_uuid()},sip_invite_domain=${domain_name},presence_id=" + name + "@${domain_name}"
}

func addBridgeEndpoint(c *Call, dialString *string, endpoint map[string]interface{}) {
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
		setSpyMap(c, getStringValueFromMap("name", endpoint, ""))

		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			*dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}
		*dialString += "user/" + getStringValueFromMap("name", endpoint, "_undef") + "@${domain_name}"

	case "user":
		setSpyMap(c, getStringValueFromMap("name", endpoint, "_undef_"))
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
