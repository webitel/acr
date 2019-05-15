/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/wlog"
	"strings"
)

const (
	multiset = "multiset"
	set      = "set"
	export   = "export"
)

func SetVar(c *Call, args interface{}) error {
	var str = ""
	var app = ""
	switch args.(type) {
	case []interface{}:
		app = multiset
		str = "^^"
		for _, v := range args.([]interface{}) {
			if s, ok := v.(string); ok && s != "" {
				s = c.router.ValidateVariable([]byte(s), true)
				str += "~" + s
				addVarIfDisconnect(c, s)
			}
		}
	case []string:
		app = multiset
		str = "^^"
		for _, v := range args.([]string) {
			if v != "" {
				v = c.router.ValidateVariable([]byte(v), true)
				str += "~" + v
				addVarIfDisconnect(c, v)
			}
		}

	case string:
		str = args.(string)
		if strings.HasPrefix(str, "all:") {
			app = export
			str = c.router.ValidateVariable([]byte(str[4:]), false)
			addVarIfDisconnect(c, str)
		} else if strings.HasPrefix(str, "nolocal:") {
			app = export
			str = c.router.ValidateVariable([]byte(str), false)
		} else if strings.HasPrefix(str, "domain:") {
			app = set
			str = c.router.ValidateVariable([]byte(str[7:]), false)
			addVarIfDisconnect(c, str)
			setDomainVariable(c, str)
		} else {
			app = set
			str = c.router.ValidateVariable([]byte(str), false)
			addVarIfDisconnect(c, str)
		}

	default:
		c.LogError("setVar", fmt.Sprintf("%v", args), "bad request")
		return nil
	}

	if c.Stopped() {
		return nil
	}

	err := c.Execute(app, str)
	if err != nil {
		c.LogError("setVar", str, err.Error())
		return err
	}
	c.LogDebug("setVar", str, "success")
	return nil
}

func addVarIfDisconnect(c *Call, v string) {
	if c.Stopped() && v != "" {
		idx := strings.Index(v, "=")
		wlog.Debug(fmt.Sprintf("call %s is disconnected setVar in LocalVariables %v", c.Id(), v))
		if idx == -1 {
			c.disconnectedVariables[v] = ""
		} else {
			c.disconnectedVariables[v[:idx]] = v[idx+1:]
		}
	}
}

func setDomainVariable(c *Call, v string) {
	if v != "" {
		idx := strings.Index(v, "=")
		wlog.Debug(fmt.Sprintf("call %s set domain variable %v", c.Id(), v))
		if idx == -1 {
			if result := <-c.router.app.Store.RouteVariables().Set(c.Domain(), v, ""); result.Err != nil {
				wlog.Error(fmt.Sprintf("call %s set domain var %v error: %s", c.Id(), v, result.Err.Error()))
			}
		} else {
			if result := <-c.router.app.Store.RouteVariables().Set(c.Domain(), v[:idx], c.ParseString(v[idx+1:])); result.Err != nil {
				wlog.Error(fmt.Sprintf("call %s set domain var %v error: %s", c.Id(), v, result.Err.Error()))
			}
		}
	} else {
		wlog.Error(fmt.Sprintf("call %s set domain variable is empty", c.Id()))
	}
}

func multiSetVar(c *Call, vars []string) (err error) {
	err = c.Execute(multiset, "^^~"+strings.Join(vars, "~"))
	return
}
