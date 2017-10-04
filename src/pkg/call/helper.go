/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/models"
	"net/url"
	"strconv"
)

func UrlEncoded(str string) string {
	u, err := url.Parse(str)
	if err != nil {
		return url.QueryEscape(str)
	}
	return u.String()
}

func getStringValueFromMap(name string, params map[string]interface{}, def string) (res string) {
	var ok bool
	var v interface{}

	if v, ok = params[name]; ok {
		switch v.(type) {
		case bool:
			if v.(bool) {
				return "true"
			} else {
				return "false"
			}
		case string:
			return v.(string)
		case int:
			return strconv.Itoa(v.(int))
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

func getArrayFromMap(arr interface{}) (res models.ArrayApplications, ok bool) {

	var tmp []interface{}
	var d models.Application
	if tmp, ok = arr.([]interface{}); ok {
		res = make(models.ArrayApplications, len(tmp))
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
	switch _args.(type) {
	case []interface{}:
		for _, e := range _args.([]interface{}) {
			if _, ok = e.(string); ok {
				argsElem = append(argsElem, c.ParseString(e.(string)))
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
