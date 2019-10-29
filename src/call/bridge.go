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

	dst := c.Destination()

	return c.Execute("bridge", fmt.Sprintf("{sip_route_uri=sip:192.168.177.9,sip_h_X-Webitel-Direction=internal,sip_h_X-Webitel-Display-Direction=inbound,sip_h_X-Webitel-User-Id=14000,sip_h_X-Webitel-Domain-Id=50}sofia/sip/%s@webitel.lo", dst))

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
		separator = "|"
	} else if strategy != "" && strategy != "multiple" {
		separator = ":_:"
	} else {
		separator = ","
	}

	dialString += "{sip_route_uri=sip:$${outbound_sip_proxy}" // + model.CALL_VARIABLE_DOMAIN_NAME + "=" + c.Domain() sip_route_uri=sip:$${outbound_sip_proxy}"

	if tmpArr, ok = getArrayStringFromMap("global", props); ok && len(tmpArr) > 0 {
		dialString += "," + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",")
	}

	if tmpArr, ok = getArrayStringFromMap("parameters", props); ok && len(tmpArr) > 0 {
		dialString += "," + strings.Join(c.router.ValidateArrayVariables(tmpArr), ",")
	}

	if tmpArr, ok = getArrayStringFromMap("codecs", props); ok && len(tmpArr) > 0 {
		dialString += ",absolute_codec_string='" + strings.Join(tmpArr, ",") + "'"
	}

	dialString += "}"

	if _, ok = props["pickup"]; ok {
		p = c.ParseString(getStringValueFromMap("pickup", props, ""))
	}

	var _endpointsStr []string
	if p != "" {
		_endpointsStr = append(_endpointsStr, "pickup/"+p+"@"+c.Domain())
	}

	if tmpArr, err = getRemoteEndpoints(c, endpoints); err != nil {
		return err
	} else {
		_endpointsStr = append(_endpointsStr, tmpArr...)
	}

	dialString += c.ParseString(strings.Join(_endpointsStr, separator))

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

func getRemoteEndpoints(call *Call, endpoints model.ArrayApplications) ([]string, error) {
	var typeName string
	request := make([]*model.EndpointsRequest, 0, 1)
	result := make([]string, 0, 1)

	for key, v := range endpoints {
		typeName = getStringValueFromMap("type", v, "")
		if typeName == "user" || typeName == "group" || typeName == "extension" {
			request = append(request, &model.EndpointsRequest{
				Key:  key,
				Type: typeName,
				Name: call.ParseString(getStringValueFromMap("name", v, "")),
			})
		}
	}

	response, err := call.router.app.GetDistinctDevices(call.DomainId(), request)
	if err != nil {
		return result, err
	}

	for key, v := range endpoints {
		typeName = getStringValueFromMap("type", v, "")

		switch typeName {
		case "user", "group", "extension":
			e := findEndpointByKey(&response, key)
			if e == nil {
				call.LogError("bridge", v, "not found response endpoint")
				continue
			}

			buildUserDialString(call, &result, e)
		default:
			fmt.Println(key, v)
		}

	}

	return result, nil
}

func findEndpointByKey(arr *[]*model.EndpointsResponse, key int) *model.EndpointsResponse {
	for _, v := range *arr {
		if v.Pos == key {
			return v
		}
	}
	return nil
}

func buildUserDialString(call *Call, result *[]string, endpoint *model.EndpointsResponse) error {

	if endpoint.Id == nil {
		*result = append(*result, "error/UNALLOCATED_NUMBER") //TODO
		return nil
	}

	variables := make([]string, 0, 1)

	variables = append(variables, fmt.Sprintf("%s=%d", model.CALL_VARIABLE_DOMAIN_ID_NAME, call.DomainId()))
	variables = append(variables, fmt.Sprintf("%s=%d", model.CALL_VARIABLE_USER_ID_NAME, *endpoint.Id))
	variables = append(variables, fmt.Sprintf("effective_callee_id_name='%v'", *endpoint.Name))
	variables = append(variables, "sip_h_X-Webitel-Direction=internal")

	if endpoint.Dnd != nil && *endpoint.Dnd {
		*result = append(*result, fmt.Sprintf("[%v]%s", strings.Join(variables, ","), "error/USER_BUSY"))
		return nil
	}

	if endpoint.Number == nil {
		*result = append(*result, fmt.Sprintf("[%v]%s", strings.Join(variables, ","), "error/UNALLOCATED_NUMBER"))
		return nil
	}

	*result = append(*result, fmt.Sprintf("[%v]sofia/sip/%s@%s", strings.Join(variables, ","), *endpoint.Number, call.Domain()))

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

	case "user":
		dialString += "[" + model.CALL_VARIABLE_DIRECTION_NAME + "=" + model.CALL_DIRECTION_INTERNAL + ","
		if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok && len(tmpArr) > 0 {
			dialString += strings.Join(c.router.ValidateArrayVariables(tmpArr), ",")
		}
		dialString += "]"

		dialString += fmt.Sprintf(model.CALL_BRIDGE_USER_TEMPLATE, getStringValueFromMap("name", endpoint, "_undef"),
			getStringValueFromMap("domainName", endpoint, "${"+model.CALL_VARIABLE_DOMAIN_NAME+"}"))
	}
	return dialString
}
