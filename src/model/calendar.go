package model

type Calendar struct {
	StartDate int64             `bson:"startDate"`
	EndDate   int64             `bson:"endDate"`
	TimeZone  map[string]string `bson:"timeZone"`
	Accept    []CalendarAccept  `bson:"accept"`
	Except    []CalendarExcept  `bson:"except"`
}

type CalendarAccept struct {
	WeekDay   int `bson:"weekDay"`
	StartTime int `bson:"startTime"`
	EndTime   int `bson:"endTime"`
}
type CalendarExcept struct {
	Name   string `bson:"name"`
	Date   int64  `bson:"date"`
	Repeat int8   `bson:"repeat"`
}
