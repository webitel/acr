package call

import (
	"fmt"
	"github.com/webitel/wlog"
)

func setProfileVariable(call *Call, varName, value string) error {
	wlog.Debug(fmt.Sprintf("call %s set profile variable %s=%s", call.Id(), varName, value))
	return call.Execute("set_profile_var", fmt.Sprintf("%s=%s", varName, value))
}
