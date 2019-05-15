/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/model"
	"time"
)

const (
	calendarStatusAhead   = "ahead"
	calendarStatusExpire  = "expire"
	calendarStatusHoliday = "holiday"
	calendarStatusInTime  = "true"
	calendarStatusOutTime = "false"
)

var weakdays = []int{7, 1, 2, 3, 4, 5, 6}

func Calendar(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, varName string
	var calendar *model.Calendar
	var extended bool

	var current, tmpDate time.Time
	var loc *time.Location
	var timestamp int64
	var currentTimeOfDay, currentWeek, currentDay, currentMonth, currentYear int

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("calendar", args, "bad request")
		return nil
	}

	if name = getStringValueFromMap("name", props, ""); name == "" {
		c.LogError("calendar", args, "name is required")
		return nil
	}

	if varName = getStringValueFromMap("setVar", props, ""); varName == "" {
		c.LogError("calendar", args, "setVar is required")
		return nil
	}

	extended = getBoolValueFromMap("extended", props, false)

	result := <-c.router.app.Store.Calendar().Get(c.Domain(), name)
	if result.Err != nil {
		c.LogError("calendar", name, result.Err.Error())
		return nil
	} else {
		calendar = result.Data.(*model.Calendar)
	}

	if c.Timezone() != "" {
		loc, _ = time.LoadLocation(c.Timezone())
	}

	if loc == nil {
		if _, ok = calendar.TimeZone["id"]; ok {
			loc, _ = time.LoadLocation(calendar.TimeZone["id"])
		}
	}

	if loc == nil {
		c.LogWarn("calendar", args, "not found timezone, use UTC")
		current = time.Now()
	} else {
		current = time.Now().In(loc)
		c.LogDebug("calendar", args, "timezone="+loc.String())
	}

	timestamp = current.UnixNano() / 1000000

	if calendar.StartDate > 0 && timestamp < calendar.StartDate {
		return callbackCalendar(c, varName, calendarStatusAhead, extended)
	} else if calendar.EndDate > 0 && timestamp > calendar.EndDate {
		return callbackCalendar(c, varName, calendarStatusExpire, extended)
	}

	ok = false
	currentWeek = getWeekday(current)
	currentTimeOfDay = current.Hour()*60 + current.Minute()

	if len(calendar.Except) > 0 {
		currentDay = current.Day()
		currentMonth = int(current.Month())
		currentYear = current.Year()

		for _, a := range calendar.Except {
			tmpDate = time.Unix(a.Date/1000, 0)
			if loc != nil {
				tmpDate.In(loc)
			}

			if tmpDate.Day() == currentDay && int(tmpDate.Month()) == currentMonth && (a.Repeat == 1 || (a.Repeat == 0 && tmpDate.Year() == currentYear)) {
				return callbackCalendar(c, varName, calendarStatusHoliday, extended)
			}
		}
	}

	if len(calendar.Accept) > 0 {
		for _, a := range calendar.Accept {
			ok = (currentWeek == a.WeekDay) && between(currentTimeOfDay, a.StartTime, a.EndTime)
			if ok {
				break
			}
		}

		if !ok {
			return callbackCalendar(c, varName, calendarStatusOutTime, extended)
		}
	}

	return callbackCalendar(c, varName, calendarStatusInTime, extended)
}

func callbackCalendar(c *Call, varName string, res string, extendsResponse bool) error {
	if extendsResponse {
		return SetVar(c, varName+"="+res)
	}

	switch res {
	case calendarStatusInTime:
		return SetVar(c, varName+"="+calendarStatusInTime)
	default:
		return SetVar(c, varName+"="+calendarStatusOutTime)
	}
}

func getWeekday(in time.Time) int {
	return weakdays[in.Weekday()]
}
