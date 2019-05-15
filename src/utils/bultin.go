package utils

import "strconv"

func ParseIntValueFromString(str string, defaultValue int) (result int) {
	var err error
	if result, err = strconv.Atoi(str); err != nil {
		result = defaultValue
	}
	return
}
