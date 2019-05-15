package mongo_store

import (
	"errors"
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/store"
	"gopkg.in/mgo.v2/bson"
)

var ErrDialerContextBadDialerId = errors.New("Bad dialer objectId")

type NoSqlOutboundQueueStore struct {
	NoSqlStore
}

func NewNoSqlOutboundQueueStore(noSqlStore NoSqlStore) store.OutboundQueueStore {
	return &NoSqlOutboundQueueStore{noSqlStore}
}

func (self NoSqlOutboundQueueStore) Exists(name, domain string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		or := []bson.M{
			{
				"name": name,
			},
		}
		if bson.IsObjectIdHex(name) {
			or = append(or, bson.M{
				"_id": bson.ObjectIdHex(name),
			})
		}

		count, err := self.GetCollection("dialer").Find(bson.M{
			"$or":    or,
			"domain": domain,
		}).Limit(1).Count()

		if err != nil {
			result.Err = err
		} else {
			result.Data = count > 0
		}
	})
}

func (self NoSqlOutboundQueueStore) ExistsMember(dialer, domain string, request *model.OutboundQueueExistsMemberRequest) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		filter := bson.M{
			"dialer": dialer,
			"domain": domain,
		}

		if request.Name != "" {
			filter["name"] = request.Name
		}

		if request.EndCause != nil {
			filter["_endCause"] = request.EndCause
		}

		if request.Communications.Type != nil {
			filter["communications.type"] = request.Communications.Type
		}

		if request.Communications.Number != "" {
			filter["communications.number"] = request.Communications.Number
		}

		if request.Communications.Description != "" {
			filter["communications.description"] = request.Communications.Description
		}

		if request.Communications.State != nil {
			filter["communications.state"] = request.Communications.State
		}

		for k, v := range request.Variables {
			filter[fmt.Sprintf("variables.%s", k)] = v
		}

		fmt.Println(filter)

		count, err := self.GetCollection("members").Find(filter).Limit(1).Count()

		if err != nil {
			result.Err = err
		} else {
			result.Data = count > 0
		}

	})
}

func (self NoSqlOutboundQueueStore) GetIVRCallFlow(id, domain string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if !bson.IsObjectIdHex(id) {
			result.Err = ErrDialerContextBadDialerId
			return
		}

		var callFlow *model.OutboundIVRCallFlow

		err := self.GetCollection("dialer").Find(bson.M{
			"_id":    bson.ObjectIdHex(id),
			"domain": domain,
		}).Select(bson.M{"_cf": 1, "amd": 1}).One(&callFlow)

		if err != nil {
			result.Err = err
		} else {
			result.Data = callFlow
		}

	})
}

func (self NoSqlOutboundQueueStore) CreateMember(member *model.OutboundQueueMember) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		result.Err = self.GetCollection("members").Insert(member)
	})
}

//TODO update fields
func (self NoSqlOutboundQueueStore) UpdateMember(id string, member *model.OutboundQueueMember) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if !bson.IsObjectIdHex(id) {
			result.Err = ErrDialerContextBadDialerId
			return
		}

		result.Err = self.GetCollection("members").UpdateId(bson.ObjectIdHex(id), bson.M{
			"$set": member,
		})
	})
}
