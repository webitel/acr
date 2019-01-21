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

const sqlCreateCallbackMemberComment = `
insert into callback_members_comment (member_id, created_by, text, created_on) 
select $1, $2, $3, $4
where exists (
 select *
 from callback_members
   inner join callback_queue c2 on callback_members.queue_id = c2.id
 where callback_members.id = $1 AND c2.domain = $5
);
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

func (db *DB) CreateCallbackMemberComment(memberId int, domainName, createdBy, text string) (error, int) {
	var id = 0
	res := db.pg.Debug().
		Raw(sqlCreateCallbackMemberComment, memberId, createdBy, text, GetMillis(), domainName)

	if res.Error == gorm.ErrRecordNotFound {
		return nil, id
	}
	if res.Error == nil {
		res.Row().Scan(&id)
	}
	return res.Error, id
}
