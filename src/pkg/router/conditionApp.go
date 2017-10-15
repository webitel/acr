/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"fmt"
	"github.com/kataras/iris/core/errors"
	"github.com/robertkrimen/otto"
	"github.com/webitel/acr/src/pkg/logger"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TimeFnList map[string]func(time.Time) string

var timeFnList TimeFnList

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

func NewConditionApplication(id string, conf AppConfig, parent *Node) *ConditionApp {
	c := &ConditionApp{}
	c.name = "if"
	c._id = id
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
	timeFnList = TimeFnList{
		"year":          getStrYear,
		"yday":          getStrYday,
		"mon":           getStrMon,
		"mday":          getStrMday,
		"week":          getStrWeek,
		"mweek":         getStrMweek,
		"wday":          getStrWday,
		"hour":          getStrHour,
		"minute":        getStrMinute,
		"minute_of_day": getStrMinOfDay,
		"time_of_day":   getStrTimeOfDay,
		"date_time":     getStrDateTime,
	}
}

func ExecTimeFn(name string, date time.Time) string {
	if fn, ok := timeFnList[name]; ok {
		return fn(date)
	}
	return ""
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
		var v otto.Value
		param := call.Argument(0).String()
		if param == "" {
			v, _ = otto.ToValue(i.Call.GetDate().Year())
		} else {
			v, _ = vm.ToValue(parseDate(param, i.Call.GetDate().Year(), 9999))
		}
		return v
	})

	sys.Set("yday", func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		param := call.Argument(0).String()
		if param == "" {
			v, _ = otto.ToValue(i.Call.GetDate().YearDay())
		} else {
			v, _ = vm.ToValue(parseDate(param, i.Call.GetDate().YearDay(), 366))
		}
		return v
	})

	sys.Set("mon", func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		param := call.Argument(0).String()

		if param == "" {
			v, _ = otto.ToValue(i.Call.GetDate().Month())
		} else {
			v, _ = vm.ToValue(parseDate(param, int(i.Call.GetDate().Month()), 12))
		}
		return v
	})

	sys.Set("mday", func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		param := call.Argument(0).String()

		if param == "" {
			v, _ = otto.ToValue(i.Call.GetDate().Day())
		} else {
			v, _ = vm.ToValue(parseDate(param, int(i.Call.GetDate().Day()), 31))
		}
		return v
	})

	sys.Set("week", func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		param := call.Argument(0).String()
		_, week := i.Call.GetDate().ISOWeek()

		if param == "" {
			v, _ = otto.ToValue(week)
		} else {
			v, _ = vm.ToValue(parseDate(param, week, 53))
		}
		return v
	})

	sys.Set("mweek", func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		param := call.Argument(0).String()

		if param == "" {
			v, _ = otto.ToValue(numberOfTheWeekInMonth(i.Call.GetDate()))
		} else {
			v, _ = vm.ToValue(parseDate(param, numberOfTheWeekInMonth(i.Call.GetDate()), 6))
		}
		return v
	})

	sys.Set("wday", func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		param := call.Argument(0).String()

		if param == "" {
			v, _ = otto.ToValue(i.Call.GetDate().Weekday() + 1)
		} else {
			v, _ = vm.ToValue(parseDate(param, getWeekday(i.Call.GetDate()), 7))
		}
		return v
	})

	sys.Set("hour", func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		param := call.Argument(0).String()

		if param == "" {
			v, _ = otto.ToValue(i.Call.GetDate().Hour())
		} else {
			v, _ = vm.ToValue(parseDate(param, int(i.Call.GetDate().Hour()), 23))
		}

		return v
	})

	sys.Set("minute", func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		param := call.Argument(0).String()

		if param == "" {
			v, _ = otto.ToValue(i.Call.GetDate().Minute())
		} else {
			v, _ = vm.ToValue(parseDate(param, int(i.Call.GetDate().Minute()), 59))
		}

		return v
	})

	sys.Set("minute_of_day", func(call otto.FunctionCall) otto.Value {
		var v otto.Value

		param := call.Argument(0).String()
		date := i.Call.GetDate()
		minOfDay := date.Hour()*60 + date.Minute()

		if param == "" {
			v, _ = otto.ToValue(minOfDay)
		} else {
			v, _ = vm.ToValue(parseDate(param, minOfDay, 1440))
		}

		return v
	})

	sys.Set("time_of_day", func(call otto.FunctionCall) (result otto.Value) {
		var tmp []string

		date := i.Call.GetDate()

		if call.Argument(0).String() == "" {
			result, _ = otto.ToValue(leadingZeros(date.Hour()) + ":" + leadingZeros(date.Minute()))
			return
		}

		current := (date.Hour() * 10000) + (int(date.Minute()) * 100) + date.Second()
		times := strings.Split(call.Argument(0).String(), ",")

		for _, v := range times {
			tmp = strings.Split(v, "-")
			if len(tmp) != 2 {
				logger.Warning("Skip parse: %v", v)
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

	sys.Set("date_time", func(call otto.FunctionCall) (result otto.Value) {
		var tmp []string
		var err error
		var t1, t2 int64

		date := i.Call.GetDate()

		if call.Argument(0).String() == "" {
			result, _ = otto.ToValue(date.Format("2006-01-02 15:04:05"))
			fmt.Println(result)
			return
		}

		currentNano := date.UnixNano()
		times := strings.Split(call.Argument(0).String(), ",")

		for _, v := range times {
			tmp = strings.Split(v, "~")
			if len(tmp) != 2 {
				logger.Warning("Skip parse: %v", v)
				continue
			}
			strings.Trim(tmp[0], tmp[0])
			strings.Trim(tmp[1], tmp[1])

			t1, err = stringDateTimeToNano(tmp[0], i.Call.GetLocation())
			if err != nil {
				logger.Error("Call %s parse date: %s", i.Call.GetUuid(), err.Error())
				continue
			}

			t2, err = stringDateTimeToNano(tmp[1], i.Call.GetLocation())
			if err != nil {
				logger.Error("Call %s parse date: %s", i.Call.GetUuid(), err.Error())
				continue
			}
			fmt.Println("%v = %v", t1, t2)

			if currentNano >= t1 && currentNano <= t2 {
				result, _ = vm.ToValue(true)
				return
			}

		}
		return
	})

	sys.Set("limit", func(call otto.FunctionCall) (result otto.Value) {
		var tmp []string
		var err error
		var i1, i2 int
		var lenParams int

		param := call.Argument(0).String()
		if param == "" {
			result, _ = otto.ToValue(false)
			return
		}

		tmp = strings.Split(param, ",")
		lenParams = len(tmp)

		if lenParams < 1 {
			result, _ = otto.ToValue(false)
			return
		}

		i1, _ = strconv.Atoi(i.Call.GetChannelVar("variable_limit_usage_" + i.Call.GetDomain() + "_" + tmp[0]))

		if lenParams > 1 {
			i2, err = strconv.Atoi(strings.Trim(tmp[1], " "))
			if err != nil {
				logger.Error("Call %s get limit: %s", i.Call.GetUuid(), err.Error())
				return
			}
			result, _ = otto.ToValue(i1 <= i2)
			return
		} else {
			result, _ = otto.ToValue(i1)
			return
		}
	})

	return sys
}

func getStrYear(date time.Time) string {
	return strconv.Itoa(date.Year())
}
func getStrYday(date time.Time) string {
	return strconv.Itoa(date.YearDay())
}
func getStrMon(date time.Time) string {
	return strconv.Itoa(int(date.Month()))
}
func getStrMday(date time.Time) string {
	return strconv.Itoa(date.Day())
}
func getStrWeek(date time.Time) string {
	_, week := date.ISOWeek()
	return strconv.Itoa(week)
}
func getStrMweek(date time.Time) string {
	return strconv.Itoa(numberOfTheWeekInMonth(date))
}
func getStrWday(date time.Time) string {
	return strconv.Itoa(getWeekday(date))
}
func getStrHour(date time.Time) string {
	return strconv.Itoa(date.Hour())
}
func getStrMinute(date time.Time) string {
	return strconv.Itoa(date.Minute())
}
func getStrMinOfDay(date time.Time) string {
	return strconv.Itoa(date.Hour()*60 + date.Minute())
}
func getStrTimeOfDay(date time.Time) string {
	return leadingZeros(date.Hour()) + ":" + leadingZeros(date.Minute())
}
func getStrDateTime(date time.Time) string {
	return date.Format("2006-01-02 15:04:05")
}

func stringDateTimeToNano(data, locationName string) (int64, error) {
	var t time.Time
	var err error
	var length = len(data)
	var loc *time.Location

	if locationName != "" {
		loc, err = time.LoadLocation(locationName)
		if err != nil {
			return 0, err
		}
	}

	if length == 19 {
		if loc != nil {
			t, err = time.ParseInLocation("2006-01-02 15:04:05", data, loc)
		} else {
			t, err = time.Parse("2006-01-02 15:04:05", data)
		}
	} else if length == 16 {
		if loc != nil {
			t, err = time.ParseInLocation("2006-01-02 15:04", data, loc)
		} else {
			t, err = time.Parse("2006-01-02 15:04", data)
		}
	} else {
		return 0, errors.New("Bad parse string:" + data)
	}

	if err != nil {
		return 0, err
	}

	return t.UnixNano(), nil
}

func leadingZeros(data int) string {
	if data < 10 {
		return "0" + strconv.Itoa(data)
	} else {
		return strconv.Itoa(data)
	}
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

var weakdays = []int{7, 1, 2, 3, 4, 5, 6}

//todo move helper (calendar use)
func getWeekday(in time.Time) int {
	return weakdays[in.Weekday()]
}
