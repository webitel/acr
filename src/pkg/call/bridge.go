/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/esl"
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
	var dialString, p, strategy, separator string
	var tmpArr []string
	var endpoints models.ArrayApplications
	var e esl.Event
	var err error

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
		separator = "+F"
	} else if strategy != "" && strategy != "multiple" {
		separator = "+A"
	} else {
		separator = "+E"
	}


	dialString += "{domain_name="+ c.Domain

	if tmpArr, ok = getArrayStringFromMap("global", props); ok && len(tmpArr) > 0 {
		dialString += "," + strings.Join(validateArrayVariables(tmpArr), ",")
	}

	if tmpArr, ok = getArrayStringFromMap("parameters", props); ok && len(tmpArr) > 0 {
		dialString += "," + strings.Join(validateArrayVariables(tmpArr), ",")
	}

	if tmpArr, ok = getArrayStringFromMap("codecs", props); ok && len(tmpArr) > 0 {
		dialString += ",absolute_codec_string='" + strings.Join(tmpArr, ",") + "'"
	}

	dialString += "}group/"

	if _, ok = props["pickup"]; ok {
		p = getStringValueFromMap("pickup", props, "")
	}

	var _endpointsStr []string
	if p != "" {
		_endpointsStr = append(_endpointsStr, UrlEncoded("pickup/" + p + "@" + c.Domain))
	}

	for _, endpoint := range endpoints {
		_endpointsStr = append(_endpointsStr, UrlEncoded(c.ParseString(addBridgeEndpoint(c, endpoint))))
	}

	dialString += strings.Join(_endpointsStr, "~") + "@" + c.Domain + separator

	e, err = c.SndMsg("bridge", dialString, true, true)
	if err != nil {
		logger.Error("Call %s bridge error: %s", c.Uuid, err)
		return err
	}

	//TODO variable_last_bridge_hangup_cause variable_bridge_hangup_cause
	if e.Header.Get("variable_bridge_hangup_cause") == "NORMAL_CLEARING" && e.Header.Get("variable_hangup_after_bridge") == "true" {
		c.SetBreak()
		//return Hangup(c, "NORMAL_CLEARING")
	}

	//TODO
	if e.Header.Get("variable_last_bridge_hangup_cause") == "ORIGINATOR_CANCEL" &&
		e.Header.Get("variable_originate_disposition") == "ORIGINATOR_CANCEL" &&
		e.Header.Get("variable_sip_redirect_dialstring") != "" &&
		e.Header.Get("variable_webitel_detect_redirect") != "false" {
		logger.Warning("Call %s detect sip redirect to %s, break this route", c.Uuid, e.Header.Get("variable_sip_redirect_dialstring"))
		c.SetBreak()
	}

	return nil
}

func addBridgeEndpoint(c *Call, endpoint map[string]interface{}) string {
	var ok bool
	var tmpArr []string
	var dialString = ""

	switch getStringValueFromMap("type", endpoint, "") {
	case "sipGateway":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}
		dialString += "sofia/gateway/" + getStringValueFromMap("name", endpoint, "_undef_") + "/" +
			getStringValueFromMap("dialString", endpoint, "_undef_")

	case "sipUri":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}
		dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "external") + "/" +
			getStringValueFromMap("dialString", endpoint, "_undef_") + "@" + getStringValueFromMap("host", endpoint, "")

	case "sipDevice":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}
		dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "internal") + "/" +
			getStringValueFromMap("name", endpoint, "_undef_") + "%" +
			getStringValueFromMap("domainName", endpoint, "") + "^" + getStringValueFromMap("dialString", endpoint, "")

	case "device":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}

		dialString += fmt.Sprintf("user/%s@%s", getStringValueFromMap("name", endpoint, "_undef"), c.Domain)

	case "user":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(validateArrayVariables(tmpArr), ",") + "]"
		}

		dialString += fmt.Sprintf("user/%s@%s", getStringValueFromMap("name", endpoint, "_undef"),
			getStringValueFromMap("domainName", endpoint, "${domain_name}"))
	}
	return dialString
}

func countSipRedirectCount(headers esl.Header) (count int) {
	for {
		if headers.Exists(fmt.Sprintf("variable_sip_redirect_contact_%d", count)) {
			count++
		} else {
			return count
		}
	}
}