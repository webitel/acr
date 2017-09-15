/**
 * Created by I. Navrotskyj on 13.09.17.
 */

package models

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

type CallFlow struct {
	Id           int               `json:"id" gorm:"column:id"`
	Debug        bool              `json:"debug" gorm:"column:debug"`
	Name         string            `json:"name" gorm:"column:name"`
	Number       string            `json:"destination_number" gorm:"column:destination_number"`
	Timezone     string            `json:"fs_timezone" gorm:"column:fs_timezone"`
	Domain       string            `json:"domain" gorm:"column:domain"`
	Callflow     ArrayApplications `json:"callflow" gorm:"column:callflow" sql:"type:json" bson:"callflow"`
	OnDisconnect ArrayApplications `json:"callflow_on_disconnect" gorm:"column:callflow_on_disconnect" bson:"onDisconnect"  sql:"type:json"`
	Version      int               `json:"version" gorm:"column:version"`
}

func (CallFlow) TableName() string {
	return "callflow_default"
}
