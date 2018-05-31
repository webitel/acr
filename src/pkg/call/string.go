/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/robertkrimen/otto"
	"github.com/webitel/acr/src/pkg/logger"
)

func String(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var vm *otto.Otto
	var varName, fnName, data, value string
	var argsElem []interface{}
	var _args interface{}

	if props, ok = args.(map[string]interface{}); ok {
		if varName = getStringValueFromMap("setVar", props, ""); varName == "" {
			logger.Error("Call %s string setVar is require", c.Uuid)
			return nil
		}

		if fnName = getStringValueFromMap("fn", props, ""); fnName == "" {
			logger.Error("Call %s string fn is require", c.Uuid)
			return nil
		}

		data = c.ParseString(getStringValueFromMap("data", props, ""))

		if fnName == "reverse" {
			value = reverse(data)
		} else if fnName == "charAt" {
			if pos := getIntValueFromMap("args", props, -1); pos > -1 {
				value = charAt(data, pos)
			}
		} else {
			if _args, ok = props["args"]; ok {
				argsElem = parseArgsToArrayInterface(c, _args)
			} else {
				argsElem = []interface{}{}
			}

			vm = otto.New()
			vm.Set("fnName", fnName)
			vm.Set("args", argsElem)
			vm.Set("data", data)
			v, err := vm.Run(`
				var value, match;

				if (args instanceof Array) {
					args = args.map(function(v) {
						if (typeof v === "string") {
							match = v.match(new RegExp('^/(.*?)/([gimy]*)$'));
							if (match) {
								return new RegExp(match[1], match[2])
							}
						}
						return v;
					})
				} else {
					args = [args]
				}

				if (typeof data[fnName] === "function") {
					value = data[fnName].apply(data, args)
				} else {
					throw "Bad string function " + fnName
				}
			`)

			if err != nil {
				logger.Error("Call %s string error: ", err.Error())
				return nil
			}

			value = v.String()
		}

		logger.Debug("Call %s string %s %s -> %s successful", c.Uuid, fnName, data, value)
		return SetVar(c, varName+"="+value)

	} else {
		logger.Error("Call %s string bad arguments %s", c.Uuid, args)
	}
	return nil
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func charAt(s string, pos int) string  {
	if len(s) > pos {
		return string(s[pos])
	}
	return ""
}