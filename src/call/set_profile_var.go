package call

import (
	"fmt"
	"github.com/webitel/wlog"
)

func setProfileVariable(call *Call, varName, value string) error {
	wlog.Debug(fmt.Sprintf("call %s set profile variable %s=%s", call.Id(), varName, value))
	return call.Execute("set_profile_var", fmt.Sprintf("%s=%s", varName, value))
}

func SetProfileVariable(call *Call, args interface{}) error {
	var err error
	switch args.(type) {
	case []interface{}:
		for _, v := range args.([]interface{}) {
			if s, ok := v.(string); ok && s != "" {
				s = call.router.ValidateVariable([]byte(s), false)
				if err = call.Execute("set_profile_var", s); err != nil {
					call.LogError("SetProfileVariable", fmt.Sprintf("%v", args), err.Error())
					return err
				}
			}
		}
	default:
		call.LogError("SetProfileVariable", args, "bad request")
	}

	return nil
}
