/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"bytes"
	"encoding/json"
	"github.com/tidwall/gjson"
	"github.com/webitel/acr/src/pkg/logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"fmt"
	"gopkg.in/xmlpath.v2"
)

func HttpRequest(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var uri string
	var err error
	var urlParam *url.URL
	var str, k, method string
	var v interface{}
	var body []byte
	var req *http.Request
	var res *http.Response
	headers := make(map[string]string)

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s httpRequest bad arguments %s", c.Uuid, args)
		return nil
	}

	if uri = getStringValueFromMap("url", props, ""); uri == "" {
		logger.Error("Call %s httpRequest url is required", c.Uuid)
		return nil
	}

	if _, ok = props["path"]; ok {
		if _, ok = props["path"].(map[string]interface{}); ok {
			for k, v = range props["path"].(map[string]interface{}) {
				str = parseMapValue(c, v)
				uri = strings.Replace(uri, "${"+k+"}", str, -1)
			}
		}
	}

	urlParam, err = url.Parse(strings.Trim(uri, " "))
	if err != nil {
		logger.Error("Call %s httpRequest parse url error: %s", c.Uuid, err.Error())
		return nil
	}

	if _, ok = props["headers"]; ok {
		if _, ok = props["headers"].(map[string]interface{}); ok {
			for k, v = range props["headers"].(map[string]interface{}) {
				headers[strings.ToLower(k)] = parseMapValue(c, v)
			}
		}
	}

	if _, ok = headers["content-type"]; !ok {
		headers["content-type"] = "application/json"
	}

	if _, ok = props["data"]; ok {

		if strings.Index(headers["content-type"],"text/xml") > -1 || strings.Index(headers["content-type"],"application/soap+xml") > -1 {
			switch props["data"].(type) {
			case string:
				body = []byte(c.ParseString(getStringValueFromMap("data", props, "")))
			}
		} else if strings.HasPrefix(headers["content-type"],"application/x-www-form-urlencoded") {
			str = ""
			switch props["data"].(type) {
			case map[string]interface{}:
				for k, v = range props["data"].(map[string]interface{}) {
					str += "&" + k + "=" + parseMapValue(c, v)
				}
				if len(str) > 0 {
					str = str[1:]
				}
			case string:
				str = props["data"].(string)
			}

			if len(str) > 0 {
				body = []byte(strings.Replace(c.ParseString(str), " ", "+", -1))
			}
		} else {
			//JSON default
			body, err = json.Marshal(props["data"])
			if err != nil {
				logger.Error("Call %s httpRequest marshal data error: %s", c.Uuid, err.Error())
				return nil
			} else {
				body = []byte(c.ParseString(string(body)))
			}
		}

	}

	method = strings.ToUpper(getStringValueFromMap("method", props, "POST"))

	req, err = http.NewRequest(method, urlParam.String(), bytes.NewBuffer(body))
	if err != nil {
		logger.Error("Call %s httpRequest create request error: %s", c.Uuid, err.Error())
		return nil
	}

	for k, str = range headers {
		req.Header.Set(k, str)
	}

	client := &http.Client{
		Timeout: time.Duration(getIntValueFromMap("timeout", props, 1000)) * time.Millisecond,
	}
	res, err = client.Do(req)
	if err != nil {
		logger.Error("Call %s httpRequest response error: %s", c.Uuid, err.Error())
		return nil
	}
	defer res.Body.Close()

	if str = getStringValueFromMap("responseCode", props, ""); str != "" {
		SetVar(c, str+"="+strconv.Itoa(res.StatusCode))
	}

	if str = getStringValueFromMap("exportCookie", props, ""); str != "" {
		if _, ok = res.Header["Set-Cookie"]; ok {
			err = SetVar(c, str+"="+strings.Join(res.Header["Set-Cookie"], ";"))
			if err != nil {
				logger.Error("Call %s httpRequest set cookie variable error: %s", c.Uuid, err.Error())
			}
		}
	}

	if res.ContentLength == 0 {
		logger.Debug("Call %s httpRequest response from %s code %v no response", c.Uuid, urlParam.String(), res.StatusCode)
		return nil
	} else {
		logger.Debug("Call %s httpRequest response from %s code %v content length %v", c.Uuid, urlParam.String(), res.StatusCode, res.ContentLength)
	}

	str = res.Header.Get("content-type")
	if strings.Index(str, "application/json") > -1 {
		if _, ok = props["exportVariables"]; ok {
			if _, ok = props["exportVariables"].(map[string]interface{}); ok {
				body, err = ioutil.ReadAll(res.Body)
				if err != nil {
					logger.Error("Call %s httpRequest read response error: %s", c.Uuid, err.Error())
					return nil
				}
				for k, v = range props["exportVariables"].(map[string]interface{}) {
					if str, ok = v.(string); ok {
						//TODO escape ?
						err = SetVar(c, "all:"+k+"="+gjson.GetBytes(body, str).String()+"")
						if err != nil {
							logger.Error("Call %s httpRequest setVat error: %s", c.Uuid, err.Error())
						}
					}
				}
			}
		}
	} else if strings.Index(str, "text/xml") > -1 {
		var xml *xmlpath.Node
		var path *xmlpath.Path

		xml, err = xmlpath.Parse(res.Body)
		if err != nil {
			logger.Error("Call %s httpRequest read XML error: %s", c.Uuid, err.Error())
			return nil
		}

		for k, v = range props["exportVariables"].(map[string]interface{}) {
			if str, ok = v.(string); ok {
				path, err = xmlpath.Compile(str)
				if err != nil {
					logger.Error("Call %s httpRequest skip xml path %s by error: %s", c.Uuid, str, err.Error())
					continue
				}

				if str, ok = path.String(xml); ok {
					err = SetVar(c, "all:"+k+"="+str)
					if err != nil {
						logger.Error("Call %s httpRequest setVat error: %s", c.Uuid, err.Error())
					}
				} else {
					logger.Debug("Call %s httpRequest not found path %s", c.Uuid, str)
				}
			}
		}

	} else {
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Error("Call %s httpRequest read response error: %s", c.Uuid, err.Error())
			return nil
		}
		fmt.Println(string(body))
		logger.Warning("Call %s httpRequest no support parse content-type %s", c.Uuid, str)
	}

	return nil
}

func parseMapValue(c *Call, v interface{}) (str string) {
	str = parseInterfaceToString(v)
	if strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
		str = c.GetChannelVar(str[2 : len(str)-1])
	}
	return str
}
