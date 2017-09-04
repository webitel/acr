/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"regexp"
	"strconv"
)

var (
	validAnswer    = regexp.MustCompile(`\b200\b|\bOK\b`)
	validPreAnswer = regexp.MustCompile(`\b183\b|\bSession Progress\b`)
	validRing      = regexp.MustCompile(`\b180\b|\bRinging\b`)
)

func Answer(c *Call, args interface{}) error {
	logger.Debug("Answer call %s", c.Uuid)
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
		logger.Error("Call %s bad answer parameters %v", c.Uuid, args)
	}

	_, err := c.SndMsg(app, "", true, true)
	if err != nil {
		logger.Error("Call %s answer error: %v", c.Uuid, err)
		return err
	}
	logger.Debug("Call %s execute %s successful", c.Uuid, app)
	return nil
}
