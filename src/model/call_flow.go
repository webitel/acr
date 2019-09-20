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

type Routing struct {
	SourceId     int               `json:"source_id" db:"source_id"`
	SourceName   string            `json:"source_name" db:"source_name"`
	SourceData   string            `json:"source_data" db:"source_data"`
	DomainId     int64             `json:"domain_id" db:"domain_id"`
	DomainName   string            `json:"domain_name" db:"domain_name"`
	Number       string            `json:"number" db:"number"`
	TimezoneId   int64             `json:"timezone_id" db:"timezone_id"`
	TimezoneName string            `json:"timezone_name" db:"timezone_name"`
	SchemeId     int64             `json:"scheme_id" db:"scheme_id"`
	SchemeName   string            `json:"scheme_name" db:"scheme_name"`
	Scheme       ArrayApplications `json:"scheme" db:"scheme"`

	Debug     bool  `json:"debug" db:"debug"`
	Variables *vars `json:"variables" db:"variables"`
}

type CallFlow struct {
	Id           int                `json:"id" db:"id"`
	Debug        *bool              `json:"debug" db:"debug"`
	Name         string             `json:"name" db:"name"`
	Number       string             `json:"destination_number" db:"destination_number"`
	Timezone     *string            `json:"fs_timezone" db:"fs_timezone"`
	Domain       string             `json:"domain" db:"domain"`
	DomainId     int64              `json:"domain_id" db:"domain_id"`
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
