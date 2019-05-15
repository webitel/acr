/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/wlog"
	"strings"
)

func Bridge(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("bridge", args, "bad request")
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

	if c.Stopped() {
		c.LogError("bridge", legA+" "+legB, "disconnected")
		return nil
	}

	legA = c.ParseString(legA)
	legB = c.ParseString(legB)

	if legA == "" || legB == "" {
		c.LogError("bridge", legA+" "+legB, "empty uuid or other_uuid")
		return nil
	}

	bridgeString := fmt.Sprintf("uuid_bridge %s %s", legA, legB)

	resBytes, err := c.Api(bridgeString)
	if err != nil {
		c.LogError("bridge", bridgeString, err.Error())
		return nil
	}
	resStr := strings.Replace(string(resBytes), "\n", "", -1)

	c.LogDebug("bridge", bridgeString, string(resStr))

	if c.Stopped() {
		return nil
	}
	return SetVar(c, resVar+"="+resStr)
}

func bridgeChannel(c *Call, props map[string]interface{}) error {
	var ok bool
	var dialString, p, strategy, separator string
	var tmpArr []string
	var endpoints model.ArrayApplications
	var err error

	if _, ok = props["endpoints"]; !ok {
		c.LogError("bridge", props, "endpoints is require")
		return nil
	}

	if endpoints, ok = getArrayFromMap(props["endpoints"]); !ok {
		c.LogError("bridge", props, "bad endpoints parameter")
		return nil
	}

	if len(endpoints) == 0 {
		c.LogError("bridge", props, "empty endpoints")
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

	dialString += "{domain_name=" + c.Domain()

	if tmpArr, ok = getArrayStringFromMap("global", props); ok && len(tmpArr) > 0 {
		dialString += "," + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",")
	}

	if tmpArr, ok = getArrayStringFromMap("parameters", props); ok && len(tmpArr) > 0 {
		dialString += "," + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",")
	}

	if tmpArr, ok = getArrayStringFromMap("codecs", props); ok && len(tmpArr) > 0 {
		dialString += ",absolute_codec_string='" + strings.Join(tmpArr, ",") + "'"
	}

	dialString += "}group/"

	if _, ok = props["pickup"]; ok {
		p = c.ParseString(getStringValueFromMap("pickup", props, ""))
	}

	var _endpointsStr []string
	if p != "" {
		_endpointsStr = append(_endpointsStr, UrlEncoded("pickup/"+p+"@"+c.Domain()))
	}

	for _, endpoint := range endpoints {
		_endpointsStr = append(_endpointsStr, UrlEncoded(c.ParseString(addBridgeEndpoint(c, endpoint))))
	}

	dialString += strings.Join(_endpointsStr, "~") + "@" + c.Domain() + separator

	err = c.Execute("bridge", dialString)
	if err != nil {
		c.LogError("bridge", dialString, err.Error())
		return err
	}

	//TODO variable_last_bridge_hangup_cause variable_bridge_hangup_cause
	if c.GetVariable("variable_bridge_hangup_cause") == "NORMAL_CLEARING" && c.GetVariable("variable_hangup_after_bridge") == "true" {
		c.SetBreak()
		//return Hangup(c, "NORMAL_CLEARING")
	}

	//TODO
	if c.GetVariable("variable_last_bridge_hangup_cause") == "ORIGINATOR_CANCEL" &&
		c.GetVariable("variable_originate_disposition") == "ORIGINATOR_CANCEL" &&
		c.GetVariable("variable_sip_redirect_dialstring") != "" &&
		c.GetVariable("variable_webitel_detect_redirect") != "false" {
		wlog.Warn(fmt.Sprintf("call %s detect sip redirect to %s, break this route", c.Id(), c.GetVariable("variable_sip_redirect_dialstring")))
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
			dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}
		dialString += "sofia/gateway/" + getStringValueFromMap("name", endpoint, "_undef_") + "/" +
			getStringValueFromMap("dialString", endpoint, "_undef_")

	case "sipUri":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}
		dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "external") + "/" +
			getStringValueFromMap("dialString", endpoint, "_undef_") + "@" + getStringValueFromMap("host", endpoint, "")

	case "sipDevice":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}
		dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "internal") + "/" +
			getStringValueFromMap("name", endpoint, "_undef_") + "%" +
			getStringValueFromMap("domainName", endpoint, "") + "^" + getStringValueFromMap("dialString", endpoint, "")

	case "device":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}

		dialString += fmt.Sprintf("user/%s@%s", getStringValueFromMap("name", endpoint, "_undef"), c.Domain)

	case "user":
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += "[" + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",") + "]"
		}

		dialString += fmt.Sprintf("user/%s@%s", getStringValueFromMap("name", endpoint, "_undef"),
			getStringValueFromMap("domainName", endpoint, "${domain_name}"))
	}
	return dialString
}
