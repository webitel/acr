package call

import (
	"github.com/webitel/acr/src/router"
	"regexp"
	"strconv"
	"strings"
)

var regCompileVar *regexp.Regexp
var regCompileGlobalVar *regexp.Regexp
var regCompileLocalRegs *regexp.Regexp
var regCompileTimeFn *regexp.Regexp
var regCompileReg *regexp.Regexp

func init() {
	regCompileReg = regexp.MustCompile(`\$(\d+)`)
	regCompileVar = regexp.MustCompile(`\$\{([\s\S]*?)\}`)
	regCompileGlobalVar = regexp.MustCompile(`\$\$\{([\s\S]*?)\}`)
	regCompileLocalRegs = regexp.MustCompile(`&reg(\d+)\.\$(\d+)`)
	regCompileTimeFn = regexp.MustCompile(`&(year|yday|mon|mday|week|mweek|wday|hour|minute|minute_of_day|time_of_day|date_time)\(\)`)
}

func (call *Call) ParseString(args string) string {
	a := regCompileGlobalVar.ReplaceAllStringFunc(args, func(varName string) string {
		return call.GetGlobalVariable(regCompileGlobalVar.FindStringSubmatch(varName)[1])
	})

	a = regCompileVar.ReplaceAllStringFunc(a, func(varName string) string {
		if strings.HasPrefix(varName, "${say_string ") || strings.HasPrefix(varName, "${hash(") ||
			strings.HasPrefix(varName, "${create_uuid(") ||
			strings.HasPrefix(varName, "${sip_authorized}") ||
			strings.HasPrefix(varName, "${verto_contact(") ||
			strings.HasPrefix(varName, "${expr(") ||
			strings.HasPrefix(varName, "${sofia_contact(") { //TODO
			return varName
		}

		t := regCompileVar.FindStringSubmatch(varName)
		if idx, err := strconv.Atoi(t[1]); err == nil {
			return call.regExp.Get("0", idx)
		}
		return call.GetVariable(t[1])
	})

	a = regCompileLocalRegs.ReplaceAllStringFunc(a, func(varName string) string {
		r := regCompileLocalRegs.FindStringSubmatch(varName)
		if len(r) == 3 {
			if values, ok := call.regExp[r[1]]; ok {
				if i, err := strconv.Atoi(r[2]); err == nil && len(values) > i {
					return values[i]
				}
			}
		}

		return ""
	})

	a = regCompileReg.ReplaceAllStringFunc(a, func(s string) string {
		r := regCompileReg.FindStringSubmatch(s)
		if len(r) == 2 {
			if idx, err := strconv.Atoi(r[1]); err == nil {
				return call.regExp.Get("0", idx)
			}
		}
		return ""
	})

	a = regCompileTimeFn.ReplaceAllStringFunc(a, func(fn string) string {
		r := regCompileTimeFn.FindStringSubmatch(fn)
		if len(r) == 2 {
			return router.ExecTimeFn(r[1], call.GetDate())
		}

		return ""
	})

	return a
}
