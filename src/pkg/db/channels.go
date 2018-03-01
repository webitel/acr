package db

func (db *DB) FindUuidByPresence (presence string) (res string)  {
	rows, err := db.pg.Debug().Table("channels").
		Select(`uuid`).
		Where(`presence_id = $1`, presence).
		Order(`created desc`, true).
		Limit(1).
		Rows()

	if err != nil {
		return ""
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&res)
	}
	return res
}