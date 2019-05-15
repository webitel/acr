/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/router"
	"github.com/webitel/wlog"
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
			c.LogError("queue", props, "bad queue name")
		}
		name += "@" + c.Domain()

		if _, ok = props["timer"]; ok {
			var arrayTimers model.ArrayApplications
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
				profile = c.Context()
			}

			if num != "" {
				c.Execute("set",
					fmt.Sprintf("transfer_after_bridge=%s:XML:%s", num, profile))
			}

		}

		err := c.Execute("callcenter", name)
		if err != nil {
			c.LogError("queue", name, err.Error())
			return err
		}

		if len(timers) > 0 {
			for _, t := range timers {
				t <- true
			}
		}

		if c.GetVariable("cc_cause") == "answered" && getStringValueFromMap("continueOnAnswered", props, "") != "true" {
			c.SetBreak()
		}
		c.CurrentQueue = ""
		c.LogDebug("queue", name, "success")

	} else {
		c.LogError("queue", args, "bad request")
	}

	return nil
}

func newTimer(c *Call, props map[string]interface{}) chan bool {
	var maxTries, tries int
	var offset, interval time.Duration
	stop := make(chan bool, 1)
	var actions model.ArrayApplications
	var iterator *router.Iterator

	var ok bool

	if _, ok = props["actions"]; ok {
		if actions, ok = getArrayFromMap(props["actions"]); ok {
			iterator = router.NewIterator("queue", actions, c)
		}
	}

	if iterator == nil {
		c.LogError("queue", props, "actions is require")
		return nil
	}

	interval = time.Duration(getIntValueFromMap("interval", props, 60))
	offset = time.Duration(getIntValueFromMap("offset", props, 0))
	maxTries = getIntValueFromMap("tries", props, 10000)
	timer := time.NewTimer(time.Second * interval)

	c.LogDebug("queue", fmt.Sprintf("interval: %d; offset: %d; tries: %d", interval, offset, maxTries), "new_timer")

	go func(uuid string) {
		defer wlog.Debug(fmt.Sprintf("call %s stop timer", uuid))
		for {
			select {
			case <-timer.C:
				tries++
				c.iterateCallApplication(iterator)
				c.LogDebug("queue", fmt.Sprintf("tries %d", tries), "")

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
	}(c.Id())
	return stop
}
