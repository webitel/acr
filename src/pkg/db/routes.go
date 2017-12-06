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
		Select(`regexp_matches($1, callflow_default.destination_number) as dest, callflow_default.id as id,
			callflow_default.destination_number as destination_number, callflow_default.name as name, callflow_default.debug as debug,
			callflow_default.domain as domain, callflow_default.fs_timezone as fs_timezone, callflow_default.callflow as callflow,
			callflow_default.callflow_on_disconnect as callflow_on_disconnect, callflow_default.version as version,
			cv.variables::JSON as variables`, destinationNumber).
		Joins("LEFT JOIN callflow_variables cv on cv.domain = callflow_default.domain").
		Where(`callflow_default.domain = $2 AND callflow_default.disabled IS NOT TRUE`, domainName).
		Order(`callflow_default."order" ASC`, true).
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
		Select(`callflow_extension.id as id, callflow_extension.destination_number as destination_number,
			callflow_extension.name as name, callflow_extension.callflow as callflow,
			callflow_extension.callflow_on_disconnect as callflow_on_disconnect, callflow_extension.version as version,
			callflow_extension.domain as domain, cv.variables::JSON as variables`).
		Joins("LEFT JOIN callflow_variables cv on cv.domain = callflow_extension.domain").
		Where(`callflow_extension.domain = $1 AND callflow_extension.destination_number = $2 `, domainName, destinationNumber).
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
		Select(`callflow_public.id as id, callflow_public.destination_number as destination_number, callflow_public.name as name,
			callflow_public.debug as debug, callflow_public.domain as domain, callflow_public.fs_timezone as fs_timezone,
			callflow_public.callflow as callflow, callflow_public.callflow_on_disconnect as callflow_on_disconnect,
			callflow_public.version as version, cv.variables::JSON as variables`).
		Joins("LEFT JOIN callflow_variables cv on cv.domain = callflow_public.domain").
		Where(`destination_number @> ARRAY[$1]::varchar[] AND disabled IS NOT TRUE`, destinationNumber).
		Limit(1).
		Scan(&def)

	if res.Error == gorm.ErrRecordNotFound {
		return def, nil
	}

	return def, res.Error
}
