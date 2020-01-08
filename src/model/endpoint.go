package model

import "github.com/lib/pq"

type EndpointsRequest struct {
	Key  int    `json:"key"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type EndpointsResponse struct {
	Pos    int     `json:"pos" db:"pos"`
	Id     *int64  `json:"id" db:"id"`
	Name   *string `json:"name" db:"name"`
	Number *string `json:"number" db:"number"`
	Dnd    *bool   `json:"dnd" db:"dnd"`
}

type Endpoint struct {
	Idx         int             `json:"idx" db:"idx"`
	TypeName    string          `json:"type_name" db:"type_name"`
	Dnd         *bool           `json:"dnd" db:"dnd"`
	Destination *string         `json:"destination" db:"destination"`
	Variables   *pq.StringArray `json:"variables" db:"variables"`
}
