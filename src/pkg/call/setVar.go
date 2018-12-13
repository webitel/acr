/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"bytes"
	"github.com/webitel/acr/src/pkg/config"
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
				s = validateVariable([]byte(s), true)
				str += "~" + s
				addVarIfDisconnect(c, s)
			}
		}
	case []string:
		app = multiset
		str = "^^"
		for _, v := range args.([]string) {
			if v != "" {
				v = validateVariable([]byte(v), true)
				str += "~" + v
				addVarIfDisconnect(c, v)
			}
		}

	case string:
		str = args.(string)
		if strings.HasPrefix(str, "all:") {
			app = export
			str = validateVariable([]byte(str[4:]), false)
			addVarIfDisconnect(c, str)
		} else if strings.HasPrefix(str, "nolocal:") {
			app = export
			str = validateVariable([]byte(str), false)
		} else if strings.HasPrefix(str, "domain:") {
			app = set
			str = validateVariable([]byte(str[7:]), false)
			addVarIfDisconnect(c, str)
			setDomainVariable(c, str)
		} else {
			app = set
			str = validateVariable([]byte(str), false)
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

func multiSetVar(c *Call, vars []string) (err error) {
	_, err = c.SndMsg(multiset, "^^~"+strings.Join(vars, "~"), true, true)
	return
}

func validateArrayVariables(vars []string) []string {
	for i := 0; i < len(vars); i++ {
		vars[i] = validateVariable([]byte(vars[i]), true)
	}
	return vars
}

func validateVariable(vars []byte, quote bool) string {
	res := bytes.SplitN(vars, []byte("="), 2)
	l := len(res)
	if l == 0 {
		return ""
	}

	if bytes.IndexAny(res[0], ",") != -1 {
		res[0] = bytes.Replace(res[0], []byte(","), []byte(""), -1)
		logger.Warning("Bad variable name replace to %s", string(res[0]))
	}

	for _, s := range protectedVariables {
		if bytes.Contains(res[0], s) {
			res[0] = append([]byte("x_"), res[0]...)
			logger.Warning("Not allow variable %s, add prefix x_", string(vars))
			break
		}
	}

	if l == 2 {
		res[1] = bytes.TrimSpace(res[1])
		if len(res[1]) == 0 && quote {
			res[1] = []byte("''")
		} else if quote {
			if !bytes.HasPrefix(res[1], []byte("'")) {
				res[1] = append([]byte("'"), res[1]...)
			}

			if !bytes.HasSuffix(res[1], []byte("'")) {
				res[1] = append(res[1], []byte("'")...)
			}
		} else {
			if bytes.HasPrefix(res[1], []byte("'")) {
				res[1] = res[1][1:]
			}

			if bytes.HasSuffix(res[1], []byte("'")) {
				res[1] = res[1][:len(res[1])-1]
			}
		}

	} else if quote {
		res = append(res, make([]byte, 2))
		res[1] = []byte("''")
	}

	return string(bytes.Join(res, []byte("=")))
}

var protectedVariables [][]byte

func init() {
	fields := config.Conf.Get("protectedVariables")
	if len(fields) > 0 {
		arr := strings.Split(fields, ",")
		protectedVariables = make([][]byte, len(arr))
		for k, v := range arr {
			protectedVariables[k] = []byte(v)
			logger.Notice("Add protected variable %s", v)
		}
	} else {
		protectedVariables = [][]byte{}
	}
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
