/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"time"
)

type AppConfig int

type Call interface {
	GetChannelVar(name string) string
	GetGlobalVar(name string) (val string)
	AddRegExp(data []string)
	GetDate() (now time.Time)
	ParseString(args string) string
	GetUuid() string
	ValidateApp(string) bool
	GetLocation() string
	GetDomain() string
}

const (
	flagBreakEnabled AppConfig = 1 << iota
	flagAsyncEnabled
	flagDumpEnabled
)

const (
	BaseAppFlag      = 1
	ConditionAppFlag = 2
)

type App interface {
	Execute(i *Iterator)
	GetName() string
	IsBreak() bool
	IsAsync() bool
	GetId() string
	GetArgs() interface{}
}

type baseApp struct {
	BaseNode
	_id    string
	name   string
	_break bool
	async  bool
	dump   bool
	tag    string
}

func NewBaseApp(name, id string) *CustomApp {
	a := &CustomApp{}
	a.name = name
	a._id = id
	return a
}

func (a *baseApp) GetId() string {
	return a._id
}

func (a *baseApp) GetArgs() interface{} {
	return nil
}

func (a *baseApp) IsBreak() bool {
	return a._break
}

func (a *baseApp) IsAsync() bool {
	return a.async
}

func (a *baseApp) setAppConfig(i AppConfig) {
	a._break = (i & flagBreakEnabled) == flagBreakEnabled
	a.async = (i & flagAsyncEnabled) == flagAsyncEnabled
	a.dump = (i & flagDumpEnabled) == flagDumpEnabled
}

func (a *baseApp) GetName() string {
	return a.name
}
