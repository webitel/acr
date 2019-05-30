/**
 * Created by I. Navrotskyj on 13.09.17.
 */

package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Application map[string]interface{}

type ArrayApplications []Application

func (j ArrayApplications) Value() (driver.Value, error) {
	str, err := json.Marshal(j)
	return string(str), err
}

func (j *ArrayApplications) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, &j)
	}
	return errors.New("Error")
}

type vars map[string]string

func (j vars) Value() (driver.Value, error) {
	str, err := json.Marshal(j)
	return string(str), err
}

func (j *vars) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, &j)
	}
	return errors.New("Error")
}

type CallFlow struct {
	Id           int                `json:"id" db:"id"`
	Debug        *bool              `json:"debug" db:"debug"`
	Name         string             `json:"name" db:"name"`
	Number       string             `json:"destination_number" db:"destination_number"`
	Timezone     *string            `json:"fs_timezone" db:"fs_timezone"`
	Domain       string             `json:"domain" db:"domain"`
	Callflow     ArrayApplications  `json:"callflow" db:"callflow" sql:"type:json" bson:"callflow"`
	OnDisconnect *ArrayApplications `json:"callflow_on_disconnect" db:"callflow_on_disconnect" bson:"onDisconnect"  sql:"type:json"`
	Version      int                `json:"version" db:"version"`
	Variables    *vars              `json:"variables" db:"variables"`
}

func (CallFlow) TableName() string {
	return "callflow_default"
}

func MapInterfaceToArrApplications(data []map[string]interface{}) ArrayApplications {
	r := make(ArrayApplications, len(data))
	for i, v := range data {
		r[i] = Application(v)
	}
	return r
}
