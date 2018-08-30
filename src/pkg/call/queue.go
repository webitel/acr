/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
	"github.com/webitel/acr/src/pkg/router"
	"regexp"
	"strings"
	"time"
)

var validQueueName = regexp.MustCompile(`^[a-zA-Z0-9+_-]+$`)

func Queue(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var name string
	timers := []chan bool{}
	var timer map[string]interface{}

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		if strings.HasPrefix(name, "${") {
			name = c.ParseString(name)
		}
		if name == "" || !validQueueName.MatchString(name) {
			logger.Error("Call %s bad queue name: %s", c.Uuid, name)
		}
		name += "@" + c.Domain

		if _, ok = props["timer"]; ok {
			var arrayTimers models.ArrayApplications
			if timer, ok = applicationToMapInterface(props["timer"]); ok {
				timers = append(timers, newTimer(c, timer))
			} else if arrayTimers, ok = getArrayFromMap(props["timer"]); ok {
				for _, timer = range arrayTimers {
					timers = append(timers, newTimer(c, timer))
				}
			}
		}

		c.CurrentQueue = name

		if _, ok = props["startPosition"]; ok {
			var startPositionVarName string
			switch props["startPosition"].(type) {
			case string:
				startPositionVarName = props["startPosition"].(string)
			case map[string]interface{}:
				startPositionVarName = getStringValueFromMap("var", props["startPosition"].(map[string]interface{}), startPositionVarName)
			}

			if startPositionVarName == "" {
				startPositionVarName = "cc_start_position"
			}
			go func(varName string) {
				time.Sleep(time.Millisecond * 500)
				QueuePosition(c, map[string]interface{}{
					"var": varName,
				})
			}(startPositionVarName)
		}

		if transferAfterBridge := getStringValueFromMap("transferAfterBridge", props, ""); transferAfterBridge != "" {
			transferAfterBridge := strings.SplitN(transferAfterBridge, ":", 2)
			var num = ""
			var profile = ""

			if len(transferAfterBridge) == 2 {
				num = transferAfterBridge[1]
				profile = transferAfterBridge[0]
			} else {
				num = transferAfterBridge[0]
				profile = c.Conn.GetContextName()
			}

			if num != "" {
				c.SndMsg("set",
					fmt.Sprintf("transfer_after_bridge=%s:XML:%s", num, profile), true, false)
			}

		}

		e, err := c.SndMsg("callcenter", name, true, true)
		if err != nil {
			logger.Error("Call %s queue error: ", c.Uuid, err)
			return err
		}

		if len(timers) > 0 {
			for _, t := range timers {
				t <- true
			}
		}

		if e.Header.Get("variable_cc_cause") == "answered" && getStringValueFromMap("continueOnAnswered", props, "") != "true" {
			c.SetBreak()
		}
		c.CurrentQueue = ""
		logger.Debug("Call %s queue %s successful", c.Uuid, name)

	} else {
		logger.Error("Call %s queue bad arguments %s", c.Uuid, args)
	}

	return nil
}

func newTimer(c *Call, props map[string]interface{}) chan bool {
	var maxTries, tries int
	var offset, interval time.Duration
	stop := make(chan bool, 1)
	var actions models.ArrayApplications
	var iterator *router.Iterator

	var ok bool

	if _, ok = props["actions"]; ok {
		if actions, ok = getArrayFromMap(props["actions"]); ok {
			iterator = router.NewIterator(actions, c)
		}
	}

	if iterator == nil {
		logger.Error("Call %s bad actions parameters", c.Uuid)
		return nil
	}

	interval = time.Duration(getIntValueFromMap("interval", props, 60))
	offset = time.Duration(getIntValueFromMap("offset", props, 0))
	maxTries = getIntValueFromMap("tries", props, 10000)
	timer := time.NewTimer(time.Second * interval)

	logger.Debug("Call %s new timer - interval: %d; offset: %d; tries: %d", c.Uuid, interval, offset, maxTries)
	go func(uuid string) {
		defer logger.Debug("Call %s stop timer", uuid)
		for {
			select {
			case <-timer.C:
				tries++
				routeCallIterator(c, iterator)
				logger.Debug("Call %s execute timer, trie: %d", uuid, tries)

				interval += offset
				if tries >= maxTries || interval < 1 {
					timer.Stop()
					return
				}
				timer = time.NewTimer(time.Second * interval)

			case <-stop:
				timer.Stop()
				return
			}
		}
	}(c.Uuid)
	return stop
}
