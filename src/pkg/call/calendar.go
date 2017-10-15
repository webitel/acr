/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"time"
)

type calendarAcceptT struct {
	WeekDay   int `bson:"weekDay"`
	StartTime int `bson:"startTime"`
	EndTime   int `bson:"endTime"`
}
type calendarExceptT struct {
	Name   string `bson:"name"`
	Date   int64  `bson:"date"`
	Repeat int8   `bson:"repeat"`
}

type calendarT struct {
	StartDate int64             `bson:"startDate"`
	EndDate   int64             `bson:"endDate"`
	TimeZone  map[string]string `bson:"timeZone"`
	Accept    []calendarAcceptT `bson:"accept"`
	Except    []calendarExceptT `bson:"except"`
}

var weakdays = []int{7,1,2,3,4,5,6}

func Calendar(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name, varName string
	var calendar calendarT

	var current, tmpDate time.Time
	var loc *time.Location
	var timestamp int64
	var currentTimeOfDay, currentWeek, currentDay, currentMonth, currentYear int

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s calendar bad arguments %v", c.Uuid, args)
		return nil
	}

	if name = getStringValueFromMap("name", props, ""); name == "" {
		logger.Error("Call %s calendar name is required", c.Uuid)
		return nil
	}

	if varName = getStringValueFromMap("setVar", props, ""); varName == "" {
		logger.Error("Call %s calendar setVar is required", c.Uuid)
		return nil
	}

	err := c.acr.GetCalendar(name, c.Domain, &calendar)
	if err != nil {
		logger.Error("Call %s calendar error: %s", c.Uuid, err.Error())
		return nil
	}

	if c.Timezone != "" {
		loc, _ = time.LoadLocation(c.Timezone)
	}

	if loc == nil {
		if _, ok = calendar.TimeZone["id"]; ok {
			loc, _ = time.LoadLocation(calendar.TimeZone["id"])
		}
	}

	if loc == nil {
		logger.Warning("Call %s calendar no found timezone use server", c.Uuid)
		current = time.Now()
	} else {
		current = time.Now().In(loc)
		logger.Debug("Call %s calendar use timezone %s", c.Uuid, loc.String())
	}

	timestamp = current.UnixNano() / 1000000

	if calendar.StartDate > 0 && timestamp < calendar.StartDate {
		return callbackCalendar(c, varName, false)
	} else if calendar.EndDate > 0 && timestamp > calendar.EndDate {
		return callbackCalendar(c, varName, false)
	}

	if len(calendar.Accept) == 0 {
		return callbackCalendar(c, varName, false)
	}

	ok = false
	currentWeek = getWeekday(current)
	currentTimeOfDay = current.Hour()*60 + current.Minute()

	for _, a := range calendar.Accept {
		ok = (currentWeek == a.WeekDay) && between(currentTimeOfDay, a.StartTime, a.EndTime)
		if ok {
			break
		}
	}

	if !ok {
		return callbackCalendar(c, varName, false)
	}

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
				return callbackCalendar(c, varName, false)
			}
		}
	}

	return callbackCalendar(c, varName, true)
}

func callbackCalendar(c *Call, varName string, res bool) error {
	var data string
	if res {
		data = "true"
	} else {
		data = "false"
	}

	return SetVar(c, varName+"="+data)
}

func getWeekday(in time.Time) int {
	return weakdays[in.Weekday()]
}
