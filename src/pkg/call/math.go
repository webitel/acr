/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/robertkrimen/otto"
	"github.com/webitel/acr/src/pkg/logger"
	"math/rand"
	"time"
)

func Math(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var fnName, setVar, value string

	var argsElem []interface{}
	var _args interface{}
	var vm *otto.Otto

	if props, ok = args.(map[string]interface{}); ok {
		if setVar = getStringValueFromMap("setVar", props, ""); setVar == "" {
			logger.Error("Call %s math setVar is required", c.Uuid)
			return nil
		}

		fnName = getStringValueFromMap("fn", props, "random")

		if _args, ok = props["data"]; ok {
			if _, ok = _args.(string); ok {
				for _, ch := range c.ParseString(_args.(string)) {
					argsElem = append(argsElem, string(ch))
				}
			} else {
				argsElem = parseArgsToArrayInterface(c, _args)
			}
		} else {
			argsElem = []interface{}{}
		}

		if fnName == "random" || fnName == "" {
			_args = random(argsElem)
		} else {
			vm = otto.New()
			vm.Set("fnName", fnName)
			vm.Set("args", argsElem)
			v, err := vm.Run(`
				var value;

				if (typeof Math[fnName] === "function") {
					value = Math[fnName].apply(null, args);
				} else if (Math.hasOwnProperty(fnName)) {
					value = Math[fnName]
				} else {
					throw "Bad Math function " + fnName
				}

				if (isNaN(value)) {
					value = ""
				}

				value += "";
			`)

			if err != nil {
				logger.Error("Call %s Math error: ", err.Error())
				return nil
			}

			_args = v.String()
		}

		value = parseInterfaceToString(_args)
		logger.Debug("Call %s math %s -> %s successful", c.Uuid, fnName, value)
		return SetVar(c, setVar+"="+value)

	} else {
		logger.Error("Call %s math bad arguments %s", c.Uuid, args)
	}
	return nil
}

func random(arr []interface{}) interface{} {
	if len(arr) == 0 {
		return ""
	}
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(arr)
	return arr[n]
}
