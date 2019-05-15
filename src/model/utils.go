package model

import "time"

func GetMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
