/**
 * Created by I. Navrotskyj on 15.09.17.
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type variable map[string]string

type DomainVariables struct {
	Id        int      `gorm:"column:id"`
	Variables variable `gorm:"column:variables"`
}

func (j variable) Value() (driver.Value, error) {
	str, err := json.Marshal(j)
	return string(str), err
}

func (j *variable) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, &j)
	}
	return errors.New("Error")
}
