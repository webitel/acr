/**
 * Created by I. Navrotskyj on 13.09.17.
 */

package db

import (
	"github.com/jinzhu/gorm"
	"github.com/webitel/acr/src/pkg/models"
)

func (db *DB) FindDefault(domainName, destinationNumber string) (models.CallFlow, error) {

	def := models.CallFlow{}
	res := db.pg.Debug().Table("callflow_default").
		Select(`regexp_matches($1, destination_number) as dest, id, destination_number, name, debug, domain, fs_timezone, callflow, callflow_on_disconnect, version`, destinationNumber).
		Where(`domain = $2 AND disabled <> TRUE`, domainName).
		Order(`"order" ASC`, true).
		Limit(1).
		Scan(&def)

	if res.Error == gorm.ErrRecordNotFound {
		return def, nil
	}

	return def, res.Error
}

func (db *DB) FindExtension(destinationNumber string, domainName string) (models.CallFlow, error) {

	def := models.CallFlow{}
	res := db.pg.Debug().Table("callflow_extension").
		Select(`id, destination_number, name, callflow, callflow_on_disconnect, version`).
		Where(`domain = $1 AND destination_number = $2 `, domainName, destinationNumber).
		Limit(1).
		Scan(&def)

	if res.Error == gorm.ErrRecordNotFound {
		return def, nil
	}

	return def, res.Error
}

func (db *DB) FindPublic(destinationNumber string) (models.CallFlow, error) {

	def := models.CallFlow{}
	res := db.pg.Debug().Table("callflow_public").
		Select(`id, destination_number, name, debug, domain, fs_timezone, callflow, callflow_on_disconnect, version`).
		Where(`destination_number @> ARRAY[$1]::varchar[] AND disabled <> TRUE`, destinationNumber).
		Limit(1).
		Scan(&def)

	if res.Error == gorm.ErrRecordNotFound {
		return def, nil
	}

	return def, res.Error
}
