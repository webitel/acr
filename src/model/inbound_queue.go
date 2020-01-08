package model

type InboundMember struct {
	CallId     string
	Name       string
	Number     string
	ProviderId string
	Priority   int
}

type MemberAttempt struct {
	QueueId   int64  `db:"queue_id"`
	Enabled   bool   `db:"enabled"`
	AttemptId *int64 `db:"attempt_id"`
}

type InboundQueueInfo struct {
	Id            int64              `json:"id" db:"id"`
	Name          string             `json:"name" db:"name"`
	Timeout       int                `json:"timeout" db:"timeout"`
	UpdatedAt     int64              `json:"updated_at" db:"updated_at"`
	MaxCalls      int                `json:"max_calls" db:"max_calls"`
	ActiveCalls   int                `json:"active_calls" db:"active_calls"`
	Enabled       bool               `json:"enabled" db:"enabled"`
	CalendarReady bool               `json:"calendar_ready" db:"calendar_ready"`
	Schema        *ArrayApplications `json:"schema" db:"schema"`
}
