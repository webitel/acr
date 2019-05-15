/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"regexp"
	"strconv"
)

var (
	validAnswer    = regexp.MustCompile(`\b200\b|\bOK\b`)
	validPreAnswer = regexp.MustCompile(`\b183\b|\bSession Progress\b`)
	validRing      = regexp.MustCompile(`\b180\b|\bRinging\b`)
)

func Answer(c *Call, args interface{}) error {
	var str string
	var app string
	switch args.(type) {
	case string:
		str = args.(string)
	case int:
		str = strconv.Itoa(args.(int))
	}

	if str == "" || validAnswer.MatchString(str) {
		app = "answer"
	} else if validPreAnswer.MatchString(str) {
		app = "pre_answer"
	} else if validRing.MatchString(str) {
		app = "ring_ready"
	} else {
		c.LogError("answer", app, "bad request")
		return nil
	}

	err := c.Execute(app, "")
	if err != nil {
		c.LogError("answer", app, err.Error())
		return err
	}
	c.LogDebug("answer", app, "success")
	return nil
}
