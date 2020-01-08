/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/router"
	"regexp"
)

var validQueueName = regexp.MustCompile(`^[a-zA-Z0-9+_-]+$`)

type queue struct {
	call     *Call
	info     *model.InboundQueueInfo
	stop     chan struct{}
	state    chan int
	iterator *router.Iterator
}

func NewQueue(call *Call, info *model.InboundQueueInfo) *queue {
	return &queue{
		call:     call,
		info:     info,
		stop:     make(chan struct{}),
		state:    make(chan int),
		iterator: nil,
	}
}

func (q *queue) Join() {
	SetVar(nil, q.call, []string{
		"cc_node_id=call-center-1",
		"grpc_originate_success=true", //TODO
		"valet_hold_music=silence",
		fmt.Sprintf("valet_parking_timeout=%d", q.info.Timeout),
		fmt.Sprintf("cc_queue_member_priority=%d", 10),
		fmt.Sprintf("cc_queue_id=%d", q.info.Id),
		fmt.Sprintf("cc_queue_updated_at=%d", q.info.UpdatedAt),
		fmt.Sprintf("cc_queue_name=%s", q.info.Name),
		fmt.Sprintf("cc_call_id=%s", model.NewId()),
	})

	if q.info.Schema != nil {
		q.iterator = router.NewIterator(q.info.Name, *q.info.Schema, q.call)
		go func() {
			q.call.iterateCallApplication(q.iterator)
		}()
	}
	q.call.Execute("answer", "")
	//q.call.Execute("echo", "1000")
	//q.call.Execute("echo", "1000")
	//return

	q.call.Execute("valet_park", fmt.Sprintf("queue_%d ${uuid}", q.info.Id))
	if q.iterator != nil {
		q.iterator.SetCancel()
	}
	close(q.stop)
}

func Queue(scope Scope, c *Call, args interface{}) error {
	var ok bool
	if _, ok = args.(map[string]interface{}); ok {

		info, err := c.router.app.Store.InboundQueue().InboundInfo(50, "Inbound")

		if err != nil {
			panic(err.Error())
		}
		//TODO add check max > active; calendar!
		// cluster name from info
		Answer(scope, c, "200")
		q := NewQueue(c, info)
		q.Join()
		c.SetBreak()

		//Hangup(c, "USER_BUSY")
		//c.Execute("park", "")
	}
	return nil
}
