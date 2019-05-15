package model

type PrivateRoute struct {
	Uuid      string            `json:"uuid" db:"uuid"`
	Domain    string            `json:"domain" db:"domain"`
	Timezone  string            `json:"fs_timezone" db:"fs_timezone"`
	Callflow  ArrayApplications `json:"callflow" db:"callflow" bson:"callflow"`
	Deadline  int               `json:"deadline" db:"deadline" bson:"deadline"`
	Variables *vars             `json:"variables" db:"variables"`
}
