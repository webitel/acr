package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/robertkrimen/otto"
	"time"
	"errors"
)

var errTimeout = errors.New("timeout")

func JavaScript(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var src string
	var setVar string

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s JavaScript bad arguments %s", c.Uuid, args)
		return nil
	}

	if src = getStringValueFromMap("data", props, ""); src == "" {
		logger.Error("Call %s JavaScript data is required", c.Uuid)
		return nil
	}

	if setVar = c.ParseString(getStringValueFromMap("setVar", props, "")); setVar == "" {
		logger.Error("Call %s JavaScript setVar is required", c.Uuid)
		return nil
	}

	defer func() {
		if caught := recover(); caught != nil {
			logger.Error("Call %s JavaScript error: %v", c.Uuid, caught)
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
		res, err := vm.ToValue(c.GetGlobalVar(call.Argument(0).String()))
		if err != nil {
			return otto.Value{}
		}
		return res
	})

	vm.Set("_getChannelVar", func(call otto.FunctionCall) otto.Value {
		res, err := vm.ToValue(c.GetChannelVar(call.Argument(0).String()))
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
		logger.Error("Call %s JavaScript error: %s", c.Uuid, err.Error())
		logger.Error("%s", src)
		return nil
	}

	return SetVar(c, setVar + "=" + result.String())
}