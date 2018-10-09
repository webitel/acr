/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"encoding/json"
	"github.com/webitel/acr/src/pkg/logger"
	"strings"

	"bytes"
)

type callData struct {
	keys   []string
	values []interface{}
}

func (s *callData) indexKey(key string) int {
	for k, v := range s.keys {
		if v == key {
			return k
		}
	}
	return -1
}

func (s *callData) Add(name string, value interface{}) {
	if idx := s.indexKey(name); idx == -1 {
		s.keys = append(s.keys, name)
		s.values = append(s.values, value)
	} else {
		s.values[idx] = value
	}
}

func (s *callData) Length() int {
	return len(s.keys)
}

func (d callData) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	if len(d.keys) == 0 {
		b.WriteString("null")
		return nil, nil
	}

	b.WriteByte('{')

	for i, v := range d.keys {
		if i > 0 {
			b.WriteByte(',')
		}

		// marshal key
		key, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteByte(':')

		// marshal value
		val, err := json.Marshal(d.values[i])
		if err != nil {
			return nil, err
		}
		b.Write(val)
	}

	b.WriteByte('}')

	return b.Bytes(), nil
}

func newCallData(l int) *callData {
	return &callData{
		keys:   make([]string, 0, l),
		values: make([]interface{}, 0, l),
	}
}

func ExportVars(c *Call, args interface{}) error {

	if data, ok := args.([]interface{}); ok {
		variables := newCallData(len(data))

		var v interface{}
		var tmp string
		for _, v = range data {
			if tmp, ok = v.(string); ok {
				variables.Add(tmp, c.Conn.ChannelData.Header.Get("variable_"+tmp))
			}
		}

		if variables.Length() > 0 {
			body, err := json.Marshal(variables)
			if err != nil {
				logger.Error("Call %s exportVars to json error: %s", err.Error())
				return nil
			}
			err = SetVar(c, "all:webitel_data="+string(body))
			if err != nil {
				logger.Error("Call %s exportVars set webitel_data error: %s", err.Error())
				return err
			}

			err = SetVar(c, "cc_export_vars=webitel_data,"+strings.Join(variables.keys, ","))
			if err != nil {
				logger.Error("Call %s exportVars set cc_export_vars error: %s", err.Error())
				return err
			}

			logger.Debug("Call %s exportVars: %v successful", c.Uuid, variables.keys)
		}
	} else {
		logger.Error("Call %s exportVars bad arguments %s", c.Uuid, args)
	}
	return nil
}
