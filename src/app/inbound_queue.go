package app

import (
	"errors"
	"github.com/webitel/acr/src/model"
)

func (app *App) DistributeMemberToInboundQueue(domainId int64, queueName string, member *model.InboundMember) (*model.MemberAttempt, error) {
	result := <-app.Store.InboundQueue().DistributeMember(domainId, queueName, member)
	if result.Err != nil {
		return nil, result.Err
	}

	if result.Data == nil {
		return nil, errors.New("Not found queue")
	}

	return result.Data.(*model.MemberAttempt), nil
}

func (app *App) CancelIfDistributingMemberInboundQueue(attemptId int64) error {
	result := <-app.Store.InboundQueue().CancelIfDistributing(attemptId)

	if result.Err != nil {
		return result.Err
	}
	return nil
}
