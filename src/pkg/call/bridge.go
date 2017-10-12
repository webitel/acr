/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
	"strings"
)

func Bridge(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var strategy, separator, dialString, pickup, p string
	var tmpArr, params []string
	var endpoints models.ArrayApplications

	if props, ok = args.(map[string]interface{}); ok {

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

		if tmpArr, ok = getArrayStringFromMap("global", props); ok {
			dialString += "<" + strings.Join(tmpArr, ",") + ">"
		}

		dialString += "{" + "domain_name=" + c.Domain

		if tmpArr, ok = getArrayStringFromMap("parameters", props); ok {
			params = tmpArr
		}

		if tmpArr, ok = getArrayStringFromMap("codecs", props); ok {
			params = append(params, "absolute_codec_string='"+strings.Join(tmpArr, ",")+"'")
		}

		if len(params) > 0 {
			dialString += "," + strings.Join(params, ",")
		}

		dialString += "}"

		for _, endpoint := range endpoints {
			switch getStringValueFromMap("type", endpoint, "") {
			case "sipGateway":
				if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok {
					dialString += "[" + strings.Join(tmpArr, ",") + "]"
				}
				dialString += "sofia/gateway/" + getStringValueFromMap("name", endpoint, "_undef_") + "/" +
					getStringValueFromMap("dialString", endpoint, "_undef_")

			case "sipUri":
				if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok {
					dialString += "[" + strings.Join(tmpArr, ",") + "]"
				}
				dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "external") + "/" +
					getStringValueFromMap("dialString", endpoint, "_undef_") + "@" + getStringValueFromMap("host", endpoint, "")

			case "sipDevice":
				if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok {
					dialString += "[" + strings.Join(tmpArr, ",") + "]"
				}
				dialString += "sofia/" + getStringValueFromMap("profile", endpoint, "external") + "/" +
					getStringValueFromMap("name", endpoint, "_undef_") + "%" +
					getStringValueFromMap("domainName", endpoint, "") + "^" + getStringValueFromMap("dialString", endpoint, "")

			case "device":
				setSpyMap(c, getStringValueFromMap("name", endpoint, ""))

				if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok {
					dialString += "[" + strings.Join(tmpArr, ",") + "]"
				}
				dialString += "user/" + getStringValueFromMap("name", endpoint, "_undef") + "@${domain_name}"

			case "user":
				setSpyMap(c, getStringValueFromMap("name", endpoint, "_undef_"))
				switch getStringValueFromMap("proto", endpoint, "") {
				case "sip":
					dialString += getProtoParameter(getStringValueFromMap("name", endpoint, "_undef_"))

					if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok {
						dialString += "," + strings.Join(tmpArr, ",")
					}
					dialString += "]${sofia_contact(*/" + getStringValueFromMap("name", endpoint, "_undef") +
						"@${domain_name})}"

				case "webrtc":
					dialString += getProtoParameter(getStringValueFromMap("name", endpoint, "_undef_"))

					if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok {
						dialString += "," + strings.Join(tmpArr, ",")
					}

					dialString += "]${verto_contact(" + getStringValueFromMap("name", endpoint, "_undef_") +
						"@${domain_name})}"

				default:
					if tmpArr, ok = getArrayStringFromMap("parameters", endpoint); ok {
						dialString += "[" + strings.Join(tmpArr, ",") + "]"
					}

					dialString += "user/" + getStringValueFromMap("name", endpoint, "_undef_") + "@" +
						getStringValueFromMap("domainName", endpoint, "${domain_name}")
				}
			}
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

	} else {
		logger.Error("Call %s bridge bad arguments %s", c.Uuid, args)
	}

	return nil
}

func setSpyMap(c *Call, name string) {
	c.SndMsg("hash", "insert/spymap/${domain_name}-"+name+"/${uuid}", true, false)
}

func getProtoParameter(name string) string {
	return "[webitel_call_uuid=${create_uuid()},sip_invite_domain=${domain_name},presence_id=" + name + "@${domain_name}"
}
