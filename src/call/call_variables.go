package call

import (
	"github.com/webitel/acr/src/model"
	"strconv"
	"strings"
)

func (call *Call) initDomainVariables() {
	var err error

	if call.callRouting.Variables != nil && len(*call.callRouting.Variables) > 0 {
		var dVarArr []interface{}
		for k, v := range *call.callRouting.Variables {
			dVarArr = append(dVarArr, k+"="+v)
		}
		err = SetVar(call.RootScope(), call, dVarArr)
		if err != nil {
			call.LogError("SetDomainVariables", "", err.Error())
		}
	}
}

func (call *Call) GetVariable(name string) (r string) {

	if strings.HasSuffix(name, "]") {
		start := strings.LastIndex(name, "[") + 1

		if idx, err := strconv.Atoi(name[start : len(name)-1]); err == nil {
			if r = call.Connection.GetVariable("variable_" + name[0:start-1]); r != "" {
				if strings.HasPrefix(r, "ARRAY::") {
					return parseFreeSwitchArray(r[7:], idx)
				} else if idx == 0 {
					return r
				}
			}
		} else {
			call.LogError("GetVariable", name, "bad array index")
		}
		return ""

	} else {
		if call.Stopped() {
			var ok bool
			if r, ok = call.disconnectedVariables[name]; ok {
				return r
			}
		}
		if r = call.Connection.GetVariable("variable_" + name); r != "" {
			return r
		} else if r = call.Connection.GetVariable(name); r != "" {
			return r
		} else if key, ok := model.MapVariables[name]; ok {
			return call.Connection.GetVariable(key)
		}
	}

	return ""
}

func (call *Call) GetGlobalVariable(name string) string {
	val, _ := call.router.globalVarsStore.GetVariable(call.NodeId(), name)
	return val
}

func parseFreeSwitchArray(data string, pos int) string {
	d := strings.Split(data, "|:")
	if len(d) > pos {
		return d[pos]
	}
	return ""
}
