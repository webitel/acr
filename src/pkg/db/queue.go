package db

func (db *DB) CountAvailableAgent(queueName string) (count int) {
	rows, err := db.pg.Debug().Table("tiers").
		Select(`count(*) as count`).
		Joins("INNER JOIN agents a ON a.name = tiers.agent").
		Where(`tiers.queue = $1 AND a.state = 'Waiting' AND a.status in ('Available', 'Available (On Demand)')
			AND a.ready_time <= extract(epoch from now() at time zone 'utc')::BIGINT AND a.last_bridge_end < extract(epoch from now() at time zone 'utc')::BIGINT - a.wrap_up_time`, queueName).
		Rows()

	if err != nil {
		return 0
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&count)
	}

	return count
}

func (db *DB) CountAvailableMembers(queueName string) (count int) {
	rows, err := db.pg.Debug().Table("members").
		Select(`count(*) as count`).
		Where(`queue = $1 AND state in ('Trying', 'Waiting')`, queueName).
		Rows()

	if err != nil {
		return 0
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&count)
	}

	return count
}