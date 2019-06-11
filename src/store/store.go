package store

import (
	"database/sql"
	"github.com/webitel/acr/src/model"
	"time"
)

type StoreResult struct {
	Data interface{}
	Err  error
}

var ERR_NO_ROWS = sql.ErrNoRows

type StoreChannel chan StoreResult

func Do(f func(result *StoreResult)) StoreChannel {
	storeChannel := make(StoreChannel, 1)
	go func() {
		result := StoreResult{}
		f(&result)
		storeChannel <- result
		close(storeChannel)
	}()
	return storeChannel
}

func Must(sc StoreChannel) interface{} {
	r := <-sc
	if r.Err != nil {

		time.Sleep(time.Second)
		panic(r.Err)
	}

	return r.Data
}

type Store interface {
	SqlStore
	NoSqlStore
}

//TODO
type SqlStore interface {
	DefaultRoute() DefaultRouteStore
	ExtensionRoute() ExtensionRouteStore
	PublicRoute() PublicRouteStore
	PrivateRoute() PrivateRouteStore
	RouteVariables() RouteVariablesStore
	InboundQueue() InboundQueueStore
	Call() CallStore
	CallbackQueue() CallbackQueueStore
	Endpoint() EndpointStore
}

//region mongodb TODO remove this store
type NoSqlStore interface {
	GeoLocation() GeoLocationStore
	Calendar() CalendarStore
	BlackList() BlackListStore
	OutboundQueue() OutboundQueueStore
	Email() EmailStore
	Media() MediaStore
}

type OutboundQueueStore interface {
	Exists(name, domain string) StoreChannel
	ExistsMember(dialer, domain string, request *model.OutboundQueueExistsMemberRequest) StoreChannel
	GetIVRCallFlow(id, domain string) StoreChannel
	CreateMember(member *model.OutboundQueueMember) StoreChannel
	UpdateMember(id string, member *model.OutboundQueueMember) StoreChannel
}

type GeoLocationStore interface {
	Find(sysLength int, numbers []string) StoreChannel
}

type EmailStore interface {
	Config(domain string) StoreChannel
}

type MediaStore interface {
	ExistsFile(name, typeFile, domain string) StoreChannel
}

type CalendarStore interface {
	Get(domain, name string) StoreChannel
}

type BlackListStore interface {
	CountNumbers(domain, name, number string) StoreChannel
}

//endregion

type DefaultRouteStore interface {
	Get(domainId int64, destination string) StoreChannel
}

type ExtensionRouteStore interface {
	Get(domain, extension string) StoreChannel
}

type PublicRouteStore interface {
	Get(destination string) StoreChannel
}

type PrivateRouteStore interface {
	Get(callId, domain string) StoreChannel
	Create(callId, domain, timeZone string, deadline int, apps model.ArrayApplications) StoreChannel
	Remove(domain, callId string) StoreChannel
}

type RouteVariablesStore interface {
	Get(domain string) StoreChannel
	Set(domain, key, value string) StoreChannel
}

type InboundQueueStore interface {
	Exists(domain, name string) StoreChannel
	CountAvailableAgent(domain, name string) StoreChannel
	CountAvailableMembers(domain, name string) StoreChannel
}

type CallStore interface {
	GetIdByPresence(presence string) StoreChannel
}

type CallbackQueueStore interface {
	Exists(domain, name string) StoreChannel
	CreateMember(domain, queue, number, widgetName string) StoreChannel
	CreateMemberComment(memberId int64, domain, createdBy, text string) StoreChannel
}

type EndpointStore interface {
	GetDistinctDevices(domainId int64, request []*model.EndpointsRequest) StoreChannel
}
