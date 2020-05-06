package model

type OutboundIVRCallFlow struct {
	Callflow []map[string]interface{} `bson:"_cf"`
	AMD      map[string]interface{}   `bson:"amd"`
}

type OutboundQueueMember struct {
	CreatedOn        int64                              `json:"createdOn" bson:"createdOn"`
	Name             string                             `json:"name"`
	Dialer           string                             `json:"dialer"`
	Domain           string                             `json:"domain"`
	Priority         int                                `json:"priority"`
	Expire           int                                `json:"expire"`
	NextCallAfterSec *int64                             `json:"_nextTryTime,omitempty" bson:"_nextTryTime,omitempty"`
	Variables        map[string]interface{}             `json:"variables"`
	Communications   []OutboundQueueMemberCommunication `json:"communications"`
}

type OutboundQueueMemberCommunication struct {
	Number      string  `json:"number"`
	Priority    int     `json:"priority"`
	Status      int     `json:"status"`
	State       int     `json:"state"`
	Type        *string `json:"type"`
	Description string  `json:"description"`
}

type OutboundQueueExistsMemberRequest struct {
	Name           string      `json:"name"`
	EndCause       interface{} `json:"end_cause"`
	Communications struct {
		Number      string      `json:"number"`
		Type        interface{} `json:"type"`
		State       interface{} `json:"state"`
		Description string      `json:"description"`
	} `json:"communications"`
	Variables map[string]interface{} `json:"variables"`
}

func (ivr *OutboundIVRCallFlow) ToCallFlow(domain string) *CallFlow {
	var r ArrayApplications
	r = append(r, MapInterfaceToArrApplications(ivr.Callflow)...)
	r = append(r, Application{"hangup": ""})

	return &CallFlow{
		Domain:   domain,
		Version:  2,
		Callflow: r,
	}
}
