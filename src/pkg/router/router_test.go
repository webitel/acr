package router

import (
	"testing"
	"gopkg.in/mgo.v2/bson"
)


func BenchmarkIterator(b *testing.B) {
	cf := []bson.M{
		{"answer": ""},
		{"if": bson.M{
			"sysExpression": "sys.getChnVar('igor') && 1 == 1",
		}},
	}
	var interfaceSlice []interface{} = make([]interface{}, len(cf))
	for i, d := range cf {
		interfaceSlice[i] = d
	}

	for i := 0; i < b.N; i++ {

		result := &CallFlow{
			Callflow: interfaceSlice,
		}
		iter := NewIterator(result)

		for {
			v:= iter.NextApp()
			if v == nil {
				break
			}
			v.Execute(iter)
		}
	}
 }
