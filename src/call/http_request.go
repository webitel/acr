/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"gopkg.in/xmlpath.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func HttpRequest(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var res *http.Response
	var str string

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("httpRequest", args, "bad request")
		return nil
	}

	req, err := buildRequest(c, props)
	if err != nil {
		c.LogError("httpRequest", args, err.Error())
		return nil
	}

	client := &http.Client{
		Timeout: time.Duration(getIntValueFromMap("timeout", props, 1000)) * time.Millisecond,
	}

	if getStringValueFromMap("insecureSkipVerify", props, "") == "true" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	res, err = client.Do(req)
	if err != nil {
		c.LogError("httpRequest", props, err.Error())
		return nil
	}
	defer res.Body.Close()

	if str := getStringValueFromMap("responseCode", props, ""); str != "" {
		SetVar(c, str+"="+strconv.Itoa(res.StatusCode))
	}

	if str = getStringValueFromMap("exportCookie", props, ""); str != "" {
		if _, ok = res.Header["Set-Cookie"]; ok {
			err = SetVar(c, str+"="+strings.Join(res.Header["Set-Cookie"], ";"))
			if err != nil {
				c.LogError("httpRequest", "exportCookie", err.Error())
			}
		}
	}

	if res.ContentLength == 0 {
		c.LogDebug("httpRequest", args, strconv.Itoa(res.StatusCode))
		return nil
	}

	if str = getStringValueFromMap("parser", props, ""); str == "" {
		str = res.Header.Get("content-type")
	}

	var exp map[string]interface{}
	if exp, ok = props["exportVariables"].(map[string]interface{}); ok {
		return parseHttpResponse(c, str, res.Body, exp)
	}

	return nil
}

func buildRequest(c *Call, props map[string]interface{}) (*http.Request, error) {
	var ok bool
	var uri string
	var err error
	var urlParam *url.URL
	var str, k, method string
	var v interface{}
	var body []byte
	var req *http.Request
	headers := make(map[string]string)

	if uri = getStringValueFromMap("url", props, ""); uri == "" {
		return nil, errors.New("url is required")
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
		return nil, err
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

		if strings.Index(headers["content-type"], "text/xml") > -1 || strings.Index(headers["content-type"], "application/soap+xml") > -1 {
			switch props["data"].(type) {
			case string:
				body = []byte(c.ParseString(getStringValueFromMap("data", props, "")))
			}
		} else if strings.HasPrefix(headers["content-type"], "application/x-www-form-urlencoded") {
			str = ""
			urlEncodeData := url.Values{}
			switch props["data"].(type) {
			case map[string]interface{}:
				for k, v = range props["data"].(map[string]interface{}) {
					urlEncodeData.Set(k, parseMapValue(c, v))
				}
				str = urlEncodeData.Encode()
			case string:
				str = props["data"].(string)
			}
			body = []byte(str)
		} else {
			//JSON default
			body, err = json.Marshal(props["data"])
			if err != nil {
				return nil, err
			} else {
				body = []byte(c.ParseString(string(body)))
			}
		}

	}

	method = strings.ToUpper(getStringValueFromMap("method", props, "POST"))

	req, err = http.NewRequest(method, urlParam.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for k, str = range headers {
		req.Header.Set(k, str)
	}
	return req, nil
}

func parseHttpResponse(c *Call, contentType string, response io.ReadCloser, exportVariables map[string]interface{}) error {
	var err error
	var body []byte

	if strings.Index(contentType, "application/json") > -1 {
		if len(exportVariables) > 0 {
			body, err = ioutil.ReadAll(response)
			if err != nil {
				return err
			}
			for k, _ := range exportVariables {
				err = SetVar(c, "all:"+k+"="+gjson.GetBytes(body, getStringValueFromMap(k, exportVariables, "")).String()+"")
				if err != nil {
					c.LogError("httpRequest", exportVariables, err.Error())
				}
			}
		}
	} else if strings.Index(contentType, "text/xml") > -1 {
		var xml *xmlpath.Node
		var path *xmlpath.Path

		if len(exportVariables) < 1 {
			return nil
		}

		xml, err = xmlpath.Parse(response)
		if err != nil {
			c.LogError("httpRequest", exportVariables, err.Error())
			return nil
		}

		for k, _ := range exportVariables {
			path, err = xmlpath.Compile(getStringValueFromMap(k, exportVariables, ""))
			if err != nil {
				c.LogError("httpRequest", k, err.Error())
				continue
			}

			if str, ok := path.String(xml); ok {
				err = SetVar(c, "all:"+k+"="+str)
				if err != nil {
					c.LogError("httpRequest", str, err.Error())
				}
			} else {
				c.LogDebug("httpRequest", exportVariables, " not found path "+str)
			}
		}

	} else {
		body, err = ioutil.ReadAll(response)
		if err != nil {
			c.LogError("httpRequest", exportVariables, err.Error())
			return nil
		}
		fmt.Println(string(body))
		c.LogWarn("httpRequest", string(body), "no support parse content-type "+contentType)
	}

	return nil
}

func parseMapValue(c *Call, v interface{}) (str string) {
	str = parseInterfaceToString(v)
	if strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
		str = c.GetVariable(str[2 : len(str)-1])
	}
	return str
}
