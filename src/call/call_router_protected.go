package call

import (
	"bytes"
	"fmt"
	"github.com/webitel/acr/src/config"
	"github.com/webitel/wlog"
	"strings"
)

func (router *CallRouterImpl) initProtectedVariables() {
	fields := config.Conf.Get("protectedVariables")
	if len(fields) > 0 {
		arr := strings.Split(fields, ",")
		router.protectedVariables = make([][]byte, len(arr))
		for k, v := range arr {
			router.protectedVariables[k] = []byte(v)
			wlog.Info(fmt.Sprintf("add call route protected variable %s", v))
		}
	} else {
		router.protectedVariables = [][]byte{}
	}
}

func (router *CallRouterImpl) ValidateArrayVariables(vars []string) []string {
	for i := 0; i < len(vars); i++ {
		vars[i] = router.ValidateVariable([]byte(vars[i]), true)
	}
	return vars
}

func (router *CallRouterImpl) ValidateVariable(vars []byte, quote bool) string {
	res := bytes.SplitN(vars, []byte("="), 2)
	l := len(res)
	if l == 0 {
		return ""
	}

	if bytes.IndexAny(res[0], ",") != -1 {
		res[0] = bytes.Replace(res[0], []byte(","), []byte(""), -1)
		wlog.Warn(fmt.Sprintf("bad variable name replace to %s", string(res[0])))
	}

	for _, s := range router.protectedVariables {
		if bytes.Contains(res[0], s) {
			res[0] = append([]byte("x_"), res[0]...)
			wlog.Warn(fmt.Sprintf("not allow variable %s, add prefix x_", string(vars)))
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
