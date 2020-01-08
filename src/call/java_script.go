package call

import (
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	"github.com/webitel/wlog"
	"time"
)

var errTimeout = errors.New("timeout")

func JavaScript(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var src string
	var setVar string

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("javaScript", args, "bad request")
		return nil
	}

	if src = getStringValueFromMap("data", props, ""); src == "" {
		c.LogError("javaScript", props, "data is required")
		return nil
	}

	if setVar = c.ParseString(getStringValueFromMap("setVar", props, "")); setVar == "" {
		c.LogError("javaScript", props, "setVar is required")
		return nil
	}

	defer func() {
		if caught := recover(); caught != nil {
			wlog.Error(fmt.Sprintf("call %s js error %s", c.Id(), caught))
		}
	}()

	src = regCompileGlobalVar.ReplaceAllStringFunc(src, func(varName string) string {
		return "_getGlobalVar('" + regCompileGlobalVar.FindStringSubmatch(varName)[1] + "')"
	})
	src = regCompileVar.ReplaceAllStringFunc(src, func(varName string) string {
		return "_getChannelVar('" + regCompileVar.FindStringSubmatch(varName)[1] + "')"
	})

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1) // The buffer prevents blocking

	vm.Set("_getGlobalVar", func(call otto.FunctionCall) otto.Value {
		res, err := vm.ToValue(c.GetGlobalVariable(call.Argument(0).String()))
		if err != nil {
			return otto.Value{}
		}
		return res
	})

	vm.Set("_getChannelVar", func(call otto.FunctionCall) otto.Value {
		res, err := vm.ToValue(c.GetVariable(call.Argument(0).String()))
		if err != nil {
			return otto.Value{}
		}
		return res
	})

	vm.Set("_LocalDateParameters", func(call otto.FunctionCall) otto.Value {
		t := c.GetDate()
		res, err := vm.ToValue([]int{t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second()})
		if err != nil {
			return otto.Value{}
		}

		return res
	})

	go func() {
		time.Sleep(2 * time.Second) // Stop after two seconds
		vm.Interrupt <- func() {
			panic(errTimeout)
		}
	}()

	result, err := vm.Run(`
		var LocalDate = function() {
			var t = _LocalDateParameters();
			return new Date(t[0], t[1], t[2], t[3], t[4], t[5])
		};
		(function(LocalDate) {` + src + `})(LocalDate)`)
	if err != nil {
		c.LogError("javaScript", src, err.Error())
		return nil
	}
	c.LogDebug("javaScript", src, result.String())
	return SetVar(scope, c, setVar+"="+result.String())
}
