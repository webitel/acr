/**
 * Created by I. Navrotskyj on 31.08.17.
 */

package db

import (
	"github.com/jinzhu/gorm"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/models"
)

var COLLECTION_DOMAIN_VARIABLES = config.Conf.Get("mongodb:variablesCollection")

func (db *DB) GetDomainVariables(domainName string) (models.DomainVariables, error) {
	data := models.DomainVariables{}
	res := db.pg.Debug().Table("callflow_variables").
		Select(`id, variables`).
		Where(`domain = $1`, domainName).
		Limit(1).
		Scan(&data)

	if res.Error == gorm.ErrRecordNotFound {
		return data, nil
	}

	return data, res.Error
}

func (db *DB) SetDomainVariable(domainName, key, value string) error {
	res := db.pg.Exec(`
	  	with upsert as (
		  update callflow_variables
		  set variables = jsonb_set(variables, ARRAY[?], ?, TRUE )
		  where domain = ?
		  returning *
		)
		INSERT INTO callflow_variables (domain, variables)
		select ?, jsonb_set('{}', ARRAY[?],  ?, TRUE )
		WHERE NOT EXISTS (SELECT * FROM upsert);
	`, key, `"`+value+`"`, domainName, domainName, key, `"`+value+`"`)

	return res.Error
}
