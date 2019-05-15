/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"github.com/robertkrimen/otto"
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
			c.LogError("math", props, "setVar is required")
			return nil
		}

		fnName = getStringValueFromMap("fn", props, "random")

		if _args, ok = props["data"]; ok {
			argsElem = parseArgsToArrayInterface(c, _args)
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
				c.LogError("math", props, err.Error())
				return nil
			}

			_args = v.String()
		}

		value = parseInterfaceToString(_args)
		c.LogDebug("math", fnName+" = "+value, "success")
		return SetVar(c, setVar+"="+value)

	} else {
		c.LogError("math", args, "bad request")
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
