/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
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
				str += "~" + s
				addVarIfDisconnect(c, s)
			}
		}
	case []string:
		app = multiset
		str = "^^"
		for _, v := range args.([]string) {
			if v != "" {
				str += "~" + v
				addVarIfDisconnect(c, v)
			}
		}

	case string:
		str = args.(string)
		if strings.HasPrefix(str, "all:") {
			app = export
			str = str[4:]
			addVarIfDisconnect(c, str)
		} else if strings.HasPrefix(str, "nolocal:") {
			app = export
		} else if strings.HasPrefix(str, "domain:") {
			app = set
			str = str[7:]
			addVarIfDisconnect(c, str)
			setDomainVariable(c, str)
		} else {
			app = set
			addVarIfDisconnect(c, str)
		}

	default:
		logger.Error("Call %s setVar must string", c.Uuid)
		return nil
	}

	if c.Conn.Disconnected {
		return nil
	}
	_, err := c.SndMsg(app, str, true, true)
	if err != nil {
		logger.Error("Call %s setVar error: %s", c.Uuid, err)
		return err
	}
	logger.Debug("Call %s setVar: %s %s successful", c.Uuid, app, str)
	return nil
}

func addVarIfDisconnect(c *Call, v string) {
	if c.Conn.Disconnected && v != "" {
		idx := strings.Index(v, "=")
		logger.Debug("Call %s is disconnected setVar in LocalVariables %v", c.Uuid, v)
		if idx == -1 {
			c.LocalVariables[v] = ""
		} else {
			c.LocalVariables[v[:idx]] = v[idx+1:]
		}
	}
}

func setDomainVariable(c *Call, v string) {
	if v != "" {
		idx := strings.Index(v, "=")
		logger.Debug("Call %s set domain variable %v", c.Uuid, v)
		if idx == -1 {
			if err := c.acr.SetDomainVariable(c.Domain, v, ""); err != nil {
				logger.Error("Call %s set domain var %v error: %s", c.Uuid, v, err.Error())
			}
		} else {
			if err := c.acr.SetDomainVariable(c.Domain, v[:idx], c.ParseString(v[idx+1:])); err != nil {
				logger.Error("Call %s set domain var %v error: %s", c.Uuid, v, err.Error())
			}
		}
	} else {
		logger.Error("Call %s set domain variable is empty", c.Uuid)
	}
}
