package router

import (
	"fmt"
	"github.com/webitel/wlog"
)

const TRIGGER_DISCONNECTED = "disconnected"

func (i *Iterator) addTrigger(triggerName string, args []interface{}) {
	switch triggerName {
	case TRIGGER_DISCONNECTED:
		i.triggers[TRIGGER_DISCONNECTED] = NewIterator(fmt.Sprintf("trigger-%s", TRIGGER_DISCONNECTED),
			ArrInterfaceToArrayApplication(args), i.Call)
	default:
		wlog.Error(fmt.Sprintf("call %s trigger %s not implemented", i.Call.Id(), triggerName))
	}
}

func (i *Iterator) TriggerIterator(name string) (*Iterator, bool) {
	if t, ok := i.triggers[name]; ok {
		return t, true
	}
	return nil, false
}
