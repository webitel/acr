/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"net/url"
	"strconv"
	"strings"
)

func UrlEncoded(str string) string {
	var res = url.Values{"": {str}}.Encode()

	if len(res) < 2 {
		return ""
	}

	return compatibleJSEncodeURIComponent(res[1:])
	//u, err := url.ParseRequestURI(str)
	//if err != nil {
	//	return compatibleJSEncodeURIComponent(url.QueryEscape(str))
	//}
	//return compatibleJSEncodeURIComponent(u.String())
}

func compatibleJSEncodeURIComponent(str string) string {
	resultStr := str
	resultStr = strings.Replace(resultStr, "+", "%20", -1)
	resultStr = strings.Replace(resultStr, "%21", "!", -1)
	//resultStr = strings.Replace(resultStr, "%27", "'", -1)
	resultStr = strings.Replace(resultStr, "%28", "(", -1)
	resultStr = strings.Replace(resultStr, "%29", ")", -1)
	resultStr = strings.Replace(resultStr, "%2A", "*", -1)
	return resultStr
}

func getStringValueFromMap(name string, params map[string]interface{}, def string) (res string) {
	var ok bool
	var v interface{}

	if v, ok = params[name]; ok {

		switch v.(type) {
		case map[string]interface{}:
		case []interface{}:
			return def

		default:
			return fmt.Sprint(v)
		}
	}

	return def
}

func getIntValueFromMap(name string, params map[string]interface{}, def int) int {
	var ok bool
	var v interface{}
	var res int

	if v, ok = params[name]; ok {
		switch v.(type) {
		case int:
			return v.(int)
		case float64:
			return int(v.(float64))
		case float32:
			return int(v.(float32))
		case string:
			var err error
			if res, err = strconv.Atoi(v.(string)); err == nil {
				return res
			}
		}
	}

	return def
}

func getBoolValueFromMap(name string, params map[string]interface{}, def bool) bool {
	var ok bool
	if _, ok = params[name]; ok {
		if _, ok = params[name].(bool); ok {
			return params[name].(bool)
		}
	}
	return def
}

func getArrayFromMap(arr interface{}) (res model.ArrayApplications, ok bool) {

	var tmp []interface{}
	var d model.Application
	if tmp, ok = arr.([]interface{}); ok {
		res = make(model.ArrayApplications, len(tmp))
		for i, v := range tmp {
			if d, ok = v.(map[string]interface{}); ok {
				res[i] = d
				//res = append(res, d)
			}
		}
		return res, true
	}

	ok = false
	return res, ok
}

func applicationToMapInterface(data interface{}) (res map[string]interface{}, ok bool) {
	var b map[string]interface{}
	res = make(map[string]interface{})
	if b, ok = data.(map[string]interface{}); ok {
		for key, val := range b {
			res[key] = val
		}
		ok = true
		return
	}
	ok = false
	return
}

func getArrayStringFromMap(name string, params map[string]interface{}) (res []string, ok bool) {
	var tmp []interface{}
	var i interface{}

	if _, ok = params[name]; !ok {
		return
	}

	if tmp, ok = params[name].([]interface{}); !ok {
		return
	}

	for _, i = range tmp {
		if _, ok = i.(string); ok {
			res = append(res, i.(string))
		}
	}
	ok = true
	return
}

func parseArgsToArrayInterface(c *Call, _args interface{}) (argsElem []interface{}) {
	var ok bool
	var str string
	switch _args.(type) {
	case []interface{}:
		for _, e := range _args.([]interface{}) {
			if str, ok = e.(string); ok {
				if !regCompileLocalRegs.MatchString(str) && regCompileReg.MatchString(str) {
					argsElem = append(argsElem, e)
				} else {
					argsElem = append(argsElem, c.ParseString(str))
				}
			} else {
				argsElem = append(argsElem, e)
			}
		}
	case string:
		argsElem = []interface{}{
			c.ParseString(_args.(string)),
		}

	default:
		argsElem = []interface{}{_args}
	}

	return
}

func parseInterfaceToString(_args interface{}) string {
	return fmt.Sprintf("%v", _args)
}

func between(x, min, max int) bool {
	return x >= min && x <= max
}

func parseEmail(parameters interface{}) string {
	var ok bool

	switch parameters.(type) {
	case string:
		return parameters.(string)

	case []interface{}:
		var email = ""
		for _, v := range parameters.([]interface{}) {
			if _, ok = v.(string); ok {
				email += "," + v.(string)
			}
		}
		if len(email) > 0 {
			email = email[1:]
		}

		if email == "" {
			email = "none"
		}
		return email

	case []string:
		return strings.Join(parameters.([]string), ",")
	}
	return "none"
}
