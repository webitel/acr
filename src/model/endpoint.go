package model

import "encoding/json"

type EndpointsRequest struct {
	Key  int    `json:"key"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type EndpointsResponse struct {
	Pos int    `json:"pos" db:"pos"`
	Raw []byte `json:"endpoints" db:"endpoints"`
}

type Endpoint struct {
	Id      *int     `json:"id" db:"id"`
	Name    string   `json:"name" db:"name"`
	Devices []Device `json:"devices" db:"devices"`
}

type Device string

func (e *EndpointsResponse) ToEndpoints() ([]*Endpoint, error) {
	var devices []*Endpoint
	err := json.Unmarshal(e.Raw, &devices)
	return devices, err
}
