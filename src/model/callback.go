/**
 * Created by I. Navrotskyj on 30.10.17.
 */

package model

import "time"

type CallbackMember struct {
	Id         int64     `json:"id" db:"id"`
	Number     string    `json:"number" db:"number"`
	QueueId    int       `json:"queue_id" db:"queue_id"`
	QueueName  string    `json:"queue_name" db:"queue_name"`
	CreatedOn  time.Time `json:"created_on" db:"created_on"`
	WidgetId   *int      `json:"widget_id" db:"widget_id"`
	WidgetName *string   `json:"widget_name" db:"widget_name"`
}

type ExistsCallbackMemberRequest struct {
	Number *string `json:"number"`
	Done   *bool   `json:"done"`
}
