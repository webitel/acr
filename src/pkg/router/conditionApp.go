/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"github.com/webitel/acr/src/pkg/logger"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ConditionApp struct {
	baseApp
	expression string
	_then      *Node
	_else      *Node
	_vm        *otto.Otto
}

func (c *ConditionApp) GetThenNode() *Node {
	c._then.setFirst()
	return c._then
}
func (c *ConditionApp) GetElseNode() *Node {
	c._else.setFirst()
	return c._else
}

func (c *ConditionApp) Execute(i *Iterator) {
	if c._vm == nil {
		c._vm = otto.New()
	}
	injectJsSysObject(c._vm, i)

	c._vm.Run(`_result = ` + c.expression)
	if value, err := c._vm.Get("_result"); err == nil {
		if boolVal, err := value.ToBoolean(); err == nil && boolVal == true {
			logger.Debug("Condition %s true", c.expression)
			i.SetRoot(c.GetThenNode())
		} else {
			logger.Debug("Condition %s false", c.expression)
			i.SetRoot(c.GetElseNode())
		}
	} else {
		fmt.Println("ERROR JS")
	}
}

func NewConditionApplication(conf AppConfig, parent *Node) *ConditionApp {
	c := &ConditionApp{}
	c.name = "if"
	c._then = NewNode(parent)
	c._else = NewNode(parent)
	c.setAppConfig(conf)
	c.setParentNode(parent)
	return c
}

var u0001 *regexp.Regexp
var regSpace *regexp.Regexp

func init() {
	u0001 = regexp.MustCompile("\u0001")
	regSpace = regexp.MustCompile(`\s`)
}

func injectJsSysObject(vm *otto.Otto, i *Iterator) *otto.Object {
	sys, _ := vm.Object("sys = {}")
	sys.Set("getChnVar", func(call otto.FunctionCall) otto.Value {
		res, err := vm.ToValue(i.Call.GetChannelVar(call.Argument(0).String()))
		if err != nil {
			return otto.Value{}
		}
		return res
	})

	sys.Set("getGlbVar", func(call otto.FunctionCall) otto.Value {
		res, err := vm.ToValue(i.Call.GetGlobalVar(call.Argument(0).String()))
		if err != nil {
			return otto.Value{}
		}
		return res
	})

	sys.Set("match", func(call otto.FunctionCall) otto.Value {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Recovered in regexp", r)
			}
		}()

		req := call.Argument(0).String()
		val := call.Argument(1).String()
		req = u0001.ReplaceAllString(strings.Trim(req, "/"), "\\")

		r := regexp.MustCompile(req)
		data := r.FindAllString(val, -1)

		if len(data) == 0 {
			return otto.Value{}
		}
		i.Call.AddRegExp(data)
		v, _ := vm.ToValue(true)
		return v
	})

	sys.Set("year", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		v, _ := vm.ToValue(parseDate(param, i.Call.GetDate().Year(), 9999))
		return v
	})

	sys.Set("yday", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		v, _ := vm.ToValue(parseDate(param, i.Call.GetDate().YearDay(), 366))
		return v
	})

	sys.Set("mon", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		v, _ := vm.ToValue(parseDate(param, int(i.Call.GetDate().Month()), 12))
		return v
	})

	sys.Set("mday", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		v, _ := vm.ToValue(parseDate(param, int(i.Call.GetDate().Day()), 31))
		return v
	})

	sys.Set("week", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		_, week := i.Call.GetDate().ISOWeek()
		v, _ := vm.ToValue(parseDate(param, week, 53))
		return v
	})

	sys.Set("mweek", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		v, _ := vm.ToValue(parseDate(param, numberOfTheWeekInMonth(i.Call.GetDate()), 6))
		return v
	})

	sys.Set("wday", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		v, _ := vm.ToValue(parseDate(param, int(i.Call.GetDate().Weekday())+1, 7))
		return v
	})

	sys.Set("hour", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		v, _ := vm.ToValue(parseDate(param, int(i.Call.GetDate().Hour()), 23))
		return v
	})

	sys.Set("minute", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		v, _ := vm.ToValue(parseDate(param, int(i.Call.GetDate().Minute()), 59))
		return v
	})

	sys.Set("minute_of_day", func(call otto.FunctionCall) otto.Value {
		param := call.Argument(0).String()
		date := i.Call.GetDate()
		v, _ := vm.ToValue(parseDate(param, date.Hour()*60+date.Minute(), 1440))
		return v
	})

	sys.Set("time_of_day", func(call otto.FunctionCall) (result otto.Value) {
		date := i.Call.GetDate()
		current := (date.Hour() * 10000) + (int(date.Minute()) * 100) + date.Second()
		times := strings.Split(call.Argument(0).String(), ",")
		var tmp []string

		for _, v := range times {
			tmp = strings.Split(v, "-")
			if len(tmp) != 2 {
				logger.Warning("Skip parse: ", v)
				continue
			}
			if current >= parseTime(tmp[0]) && current <= parseTime(tmp[1]) {
				result, _ = vm.ToValue(true)
				return
			}
		}
		result, _ = vm.ToValue(false)
		return
	})

	sys.Set("limit", func(call otto.FunctionCall) (result otto.Value) {
		//TODO
		return
	})

	return sys
}

func parseTime(str string) (result int) {
	var err error
	var tmp int

	for i, v := range strings.Split(str, ":") {
		tmp, err = strconv.Atoi(strings.Trim(v, ` `))
		if err != nil {
			logger.Error("Bad parse time: ", err)
			return
		}
		if i == 0 {
			result += (tmp * 10000)
		} else if i == 1 {
			result += (tmp * 100)
		} else {
			result += tmp
		}
	}
	return
}

func numberOfTheWeekInMonth(now time.Time) int {
	beginningOfTheMonth := time.Date(now.Year(), now.Month(), 1, 1, 1, 1, 1, time.UTC)
	_, thisWeek := now.ISOWeek()
	_, beginningWeek := beginningOfTheMonth.ISOWeek()
	return 1 + thisWeek - beginningWeek
}

func parseDate(params string, datetime int, maxVal int) (result bool) {
	rows := strings.Split(regSpace.ReplaceAllString(params, ""), ",")

	if len(rows) == 0 {
		logger.Warning("Bad parameters: " + params)
		return
	}

	for _, v := range rows {
		if strings.Index(v, "-") != -1 {
			result = equalsDateTimeRange(datetime, strings.Trim(v, ` `), maxVal)
		} else {
			if i, err := strconv.Atoi(strings.Trim(v, ` `)); err == nil {
				result = i == datetime
			}
		}

		if result {
			return
		}
	}

	return
}

func equalsDateTimeRange(datetime int, strRange string, maxVal int) (result bool) {
	var min, max int
	var err error

	rows := strings.Split(strRange, "-")
	min, err = strconv.Atoi(rows[0])
	if err != nil {
		logger.Error("Bad parse date: ", err)
		return
	}

	if len(rows) >= 2 {
		max, err = strconv.Atoi(rows[1])
		if err != nil {
			logger.Error("Bad parse date: ", err)
			return
		}
	} else {
		max = maxVal
	}

	if min > max {
		tmp := min
		min = max
		max = tmp
	}

	result = datetime >= min && datetime <= max
	return
}
