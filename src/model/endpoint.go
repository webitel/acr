package model

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
