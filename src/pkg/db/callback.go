/**
 * Created by I. Navrotskyj on 30.10.17.
 */

package db

import (
	"github.com/jinzhu/gorm"
)

const sqlCreateCallbackMember = `
INSERT INTO callback_members (domain, queue_id, number, widget_id)
VALUES (?, (SELECT id FROM callback_queue WHERE name = ? AND domain = ? LIMIT 1), ?, (SELECT id
																					 FROM widget WHERE name = ? AND domain = ? LIMIT 1) )
returning id;
`

func (db *DB) CreateCallbackMember(domainName, queueName, number, widgetName string) (error, int) {
	var id = 0
	res := db.pg.Debug().
		Raw(sqlCreateCallbackMember, domainName, queueName, domainName, number, widgetName, domainName)

	if res.Error == gorm.ErrRecordNotFound {
		return nil, id
	}
	res.Row().Scan(&id)
	return res.Error, id
}
