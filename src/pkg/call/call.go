/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/models"
	"github.com/webitel/acr/src/pkg/router"
	"github.com/webitel/acr/src/pkg/rpc"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var regCompileVar *regexp.Regexp
var regCompileGlobalVar *regexp.Regexp
var regCompileLocalRegs *regexp.Regexp
var regCompileTimeFn *regexp.Regexp
var regCompileReg *regexp.Regexp

type Applications map[string]func(c *Call, args interface{}) error

type IBridge interface {
	GetGlobalVar(call *Call, varName string) (val string, ok bool)
	CheckBlackList(domainName, name, number string) (error, int)
	GetEmailConfig(domainName string, dataStructure interface{}) error
	GetCalendar(name, domainName string, dataStructure interface{}) error
	FindLocation(sysLength int, numbers []string, dataStructure interface{}) error
	GetDomainVariables(domainName string) (models.DomainVariables, error)
	SetDomainVariable(domainName, key, value string) error
	GetRPCCommandsQueueName() string
	AddRPCCommands(uuid string) rpc.ApiT
	FireRPCEventToEngine(rk string, option rpc.PublishingOption) error
	FireRPCEventToStorage(rk string, option rpc.PublishingOption) error
	AddMember(data interface{}) error
	UpdateMember(id string, data interface{}) error
	AddCallbackMember(domainName, queueName, number, widgetName string) (error, int)
	GetPrivateCallFlow(uuid string, domain string) (models.CallFlow, error)
	InsertPrivateCallFlow(uuid, domain, timeZone string, deadline int, apps models.ArrayApplications) error
	RemovePrivateCallFlow(uuid, domain string) error
	ExistsMediaFile(name, typeFile, domainName string) bool
	ExistsDialer(name string, domain string) bool
	ExistsMemberInDialer(dialer string, domain string, data []byte) bool
	ExistsQueue(name, domain string) bool
	FindUuidByPresence(presence string) string
	CountAvailableAgent(queueName string) (count int)
	CountAvailableMembers(queueName string) (count int)
}

var applications Applications

type ContextId int

const (
	CONTEXT_PUBLIC ContextId = 1 << iota
	CONTEXT_DEFAULT
	CONTEXT_DIALER
	CONTEXT_PRIVATE
)

type regExp map[string][]string

type Call struct {
	routeId              int
	Uuid                 string
	SwitchId             string
	Domain               string
	Timezone             string
	CurrentQueue         string
	DestinationNumber    string
	Iterator             *router.Iterator
	OnDisconnectIterator *router.Iterator
	LocalVariables       map[string]string
	RegExp               regExp
	Conn                 *esl.SConn
	breakCall            bool
	debug                bool
	debugLog             bool
	debugMap             map[string]interface{}
	acr                  IBridge
	context              ContextId
}

func (r regExp) Get(position string, idx int) string {
	if v, ok := r[position]; ok {
		if len(v) > idx {
			return v[idx]
		}
	}

	return ""
}

type domainVariablesT struct {
	Variables map[string]string `bson:"variables"`
}

func init() {
	regCompileReg = regexp.MustCompile(`\$(\d+)`)
	regCompileVar = regexp.MustCompile(`\$\{([\s\S]*?)\}`)
	regCompileGlobalVar = regexp.MustCompile(`\$\$\{([\s\S]*?)\}`)
	regCompileLocalRegs = regexp.MustCompile(`&reg(\d+)\.\$(\d+)`)
	regCompileTimeFn = regexp.MustCompile(`&(year|yday|mon|mday|week|mweek|wday|hour|minute|minute_of_day|time_of_day|date_time)\(\)`)

	applications = Applications{
		"answer":        Answer,          //1
		"hangup":        Hangup,          //2
		"setVar":        SetVar,          //3
		"goto":          GoTo,            //4
		"log":           Log,             //5
		"echo":          Echo,            //6
		"park":          Park,            //7
		"sleep":         Sleep,           //8
		"recordFile":    RecordFile,      //9
		"recordSession": RecordSession,   //10
		"execute":       ExecuteFunction, //11
		"script":        Script,          //12
		"conference":    Conference,      //13
		"schedule":      Schedule,        //14
		"playback":      Playback,        //15
		"queue":         Queue,           //16
		"ccPosition":    QueuePosition,   //17
		"setArray":      SetArray,        //18
		"break":         Break,           //19
		"exportVars":    ExportVars,      //20
		"ivr":           IVR,             //21
		"unSet":         UnSet,           //22
		"voicemail":     VoiceMail,       //23
		"agent":         Agent,           //24
		"amd":           AMD,             //25
		"blackList":     BlackList,       //26
		"bridge":        Bridge,          //27
		"string":        String,          //28
		"math":          Math,            //29
		"sendEmail":     SendEmail,       //30
		"inBandDTMF":    InBandDTMF,      //31
		"flushDTMF":     FlushDTMF,       //32
		"eavesdrop":     Eavesdrop,       //33
		"receiveFax":    ReceiveFax,      //34
		"pickup":        Pickup,          //35
		"sipRedirect":   SipRedirect,     //36
		"ringback":      RingBack,        //37
		"setSounds":     SetSounds,       //38
		"httpRequest":   HttpRequest,     //39
		"userData":      UserData,        //40
		"calendar":      Calendar,        //41
		"sendSms":       SendSMS,         //42
		"geoLocation":   GeoLocation,     //43
		"tts":           TTS,             //44
		"stt":           STT,             //45
		"member":        Member,          //46
		"limit":         Limit,           //47
		"callback":      CallbackQueue,   //48
		"cdr":           CDR,             //49
		"sendEvent":     SendEvent,       //50
		"originate":     Originate,       //51
		"exists":        Exists,          //52
		"queueStatus":   QueueStatus,     //53
		"js":            JavaScript,      //54
		"findUser":      FindUser,        //55
		"setUser":       SetUser,         //56
		//"stream":        Stream,
	}

	var tmp string
	var i = 0
	for tmp, _ = range applications {
		i++
		logger.Info("Register application %v (%v)", tmp, i)
	}
}

func (c *Call) GetDomain() string {
	return c.Domain
}

func (c *Call) GetRouteId() string {
	return strconv.Itoa(c.routeId)
}

func (c *Call) IsDebugLog() bool {
	return c.debugLog
}

func (c *Call) AddRegExp(data []string) {
	c.RegExp["reg_"+string(len(c.RegExp))] = data
}

func (c *Call) SetBreak() {
	if c.GetBreak() {
		return
	}
	logger.Debug("Call %s set break route", c.Uuid)
	c.breakCall = true
}

func (c *Call) GetBreak() bool {
	return c.breakCall
}

func (c *Call) GetDate() (now time.Time) {
	if c.Timezone != "" {
		if loc, err := time.LoadLocation(c.Timezone); err == nil {
			now = time.Now().In(loc)
			return
		}
	}

	now = time.Now()
	return
}

func (c *Call) GetLocation() string {
	return c.Timezone
}

func (c *Call) GetGlobalVar(name string) (val string) {
	val, _ = c.acr.GetGlobalVar(c, name)
	return val
}

func (c *Call) SndMsg(app string, args string, look bool, dump bool) (msg esl.Event, err error) {

	defer func(uuid string) {
		if r := recover(); r != nil {
			logger.Error("Call %s recovered in %v", uuid, r)
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}(c.Uuid)

	args = c.ParseString(args)
	logger.Notice("Execute %s -> %s", app, args)
	msg, err = c.Conn.SndMsg(app, args, look, dump)
	return msg, err
}

func parseFreeSwitchArray(data string, pos int) string {
	d := strings.Split(data, "|:")
	if len(d) > pos {
		return d[pos]
	}
	return ""
}

func (c *Call) GetChannelVar(name string) (r string) {
	if strings.HasSuffix(name, "]") {
		start := strings.LastIndex(name, "[") + 1

		if idx, err := strconv.Atoi(name[start : len(name)-1]); err == nil {
			if r = c.Conn.ChannelData.Header.Get("variable_" + name[0:start-1]); r != "" {
				if strings.HasPrefix(r, "ARRAY::") {
					return parseFreeSwitchArray(r[7:], idx)
				} else if idx == 0 {
					return r
				}
			}
		} else {
			logger.Error("Call %s bad get array index %s", c.Uuid, name)
		}
		return ""

	} else {
		if c.Conn.Disconnected {
			var ok bool
			if r, ok = c.GetLocalVariable(name); ok {
				return r
			}
		}
		if r = c.Conn.ChannelData.Header.Get("variable_" + name); r != "" {
			return r
		} else if r = c.Conn.ChannelData.Header.Get(name); r != "" {
			return r
		} else if key, ok := MapVariables[name]; ok {
			return c.Conn.ChannelData.Header.Get(key)
		}
	}

	return ""
}

func (c *Call) GetLocalVariable(name string) (string, bool) {
	v, ok := c.LocalVariables[name]
	return v, ok
}

func (c *Call) ParseString(args string) string {

	a := regCompileGlobalVar.ReplaceAllStringFunc(args, func(varName string) string {
		return c.GetGlobalVar(regCompileGlobalVar.FindStringSubmatch(varName)[1])
	})

	a = regCompileVar.ReplaceAllStringFunc(a, func(varName string) string {
		if strings.HasPrefix(varName, "${say_string ") || strings.HasPrefix(varName, "${hash(") ||
			strings.HasPrefix(varName, "${create_uuid(") ||
			strings.HasPrefix(varName, "${sip_authorized}") ||
			strings.HasPrefix(varName, "${verto_contact(") ||
			strings.HasPrefix(varName, "${expr(") ||
			strings.HasPrefix(varName, "${sofia_contact(") { //TODO
			return varName
		}
		t := regCompileVar.FindStringSubmatch(varName)
		if idx, err := strconv.Atoi(t[1]); err == nil {
			return c.RegExp.Get("0", idx)
		}
		return c.GetChannelVar(t[1])
	})

	a = regCompileLocalRegs.ReplaceAllStringFunc(a, func(varName string) string {
		r := regCompileLocalRegs.FindStringSubmatch(varName)
		if len(r) == 3 {
			if values, ok := c.RegExp[r[1]]; ok {
				if i, err := strconv.Atoi(r[2]); err == nil && len(values) > i {
					return values[i]
				}
			}
		}

		return ""
	})

	a = regCompileReg.ReplaceAllStringFunc(a, func(s string) string {
		r := regCompileReg.FindStringSubmatch(s)
		if len(r) == 2 {
			if idx, err := strconv.Atoi(r[1]); err == nil {
				return c.RegExp.Get("0", idx)
			}
		}
		return ""
	})

	a = regCompileTimeFn.ReplaceAllStringFunc(a, func(fn string) string {
		r := regCompileTimeFn.FindStringSubmatch(fn)
		if len(r) == 2 {
			return router.ExecTimeFn(r[1], c.GetDate())
		}

		return ""
	})

	return a
}

func (c *Call) GetUuid() string {
	return c.Uuid
}

func (c *Call) ValidateApp(name string) (ok bool) {
	_, ok = applications[name]
	return
}

func MakeCall(destinationNumber string, c *esl.SConn, cf *models.CallFlow, acr IBridge, context ContextId) *Call {

	call := &Call{
		routeId:           cf.Id,
		Timezone:          cf.Timezone,
		Uuid:              c.Uuid,
		Domain:            cf.Domain,
		Conn:              c,
		acr:               acr,
		LocalVariables:    make(map[string]string),
		debugMap:          make(map[string]interface{}),
		SwitchId:          c.ChannelData.Header.Get("Core-UUID"),
		DestinationNumber: destinationNumber,
		debugLog:          cf.Debug,
		RegExp:            setupNumber(cf.Number, destinationNumber),
		context:           context,
	}

	if c.ChannelData.Header.Get("variable_webitel_debug_acr") == "true" {
		call.debug = true
	}

	SetVar(call, []string{
		fmt.Sprintf("webitel_acr_schema_id=%d", cf.Id),
		fmt.Sprintf("webitel_acr_schema_name=%s", cf.Name),
	})

	if call.debug {
		call.debugMap["action"] = "execute"
		call.debugMap["uuid"] = call.Uuid
		call.debugMap["domain"] = call.Domain
	}

	call.Iterator = router.NewIterator(cf.Callflow, call)
	if len(cf.OnDisconnect) > 0 {
		call.OnDisconnectIterator = router.NewIterator(cf.OnDisconnect, call)
	} else {
		call.OnDisconnectIterator = nil
	}

	go setupDomainVariables(call, cf.Variables)

	return call
}

func setupDomainVariables(call *Call, variables map[string]string) {
	var err error

	if call.GetChannelVar("presence_data") == "" {
		SetVar(call, "presence_data="+call.Domain)
	}

	if len(variables) > 0 {
		var dVarArr []interface{}
		for k, v := range variables {
			dVarArr = append(dVarArr, k+"="+v)
		}
		err = SetVar(call, dVarArr)
		if err != nil {
			logger.Error("Call %s set domain variables error: %s", call.Uuid, err.Error())
		}
	}

	routeIterator(call)
}

func setupNumber(reg, dest string) map[string][]string {
	storage := make(map[string][]string)
	if reg != "" {
		re, err := regexp.Compile(reg)
		if err == nil {
			d := re.FindStringSubmatch(dest)
			storage["0"] = d
		}
	} else {
		storage["0"] = []string{}
	}
	return storage
}

func routeIterator(call *Call) {
	defer func(uuid string) {
		if r := recover(); r != nil {
			logger.Error("Call %s recovered in %v", uuid, r)
			switch x := r.(type) {
			case string:
				logger.Error("Call %s error: %v", call.Uuid, x)
			case error:
				logger.Error("Call %s error: %v", call.Uuid, x.Error())
			default:
				logger.Error("Call %s error: Unknown panicv", call.Uuid)
			}
		}
	}(call.Uuid)

	for {
		if call.Conn.GetDisconnected() {
			logger.Debug("Call %s disconnected", call.GetUuid())
			break
		}

		if call.GetBreak() {
			logger.Debug("Call %s break route", call.GetUuid())
			break
		}
		v := call.Iterator.NextApp()
		if v == nil {
			break
		}

		if fn, ok := applications[v.GetName()]; ok {
			if call.debug {
				call.FireDebugApplication(v)
			}

			if fn(call, v.GetArgs()) != nil {
				logger.Debug("Call %s stop connection", call.GetUuid())
				break
			}

			if v.IsBreak() {
				call.SetBreak()
				break
			}
			continue
		}
		if call.debug {
			call.FireDebugApplication(v)
		}
		v.Execute(call.Iterator)
	}

	if call.debug {
		call.FireDebugApplication(router.NewBaseApp("disconnect", "end"))
	}

	//call.Conn.Close()
}

func (c *Call) FireDebugApplication(a router.App) {
	if a.GetId() != "" {
		c.debugMap["app-id"] = a.GetId()
		c.debugMap["app-name"] = a.GetName()

		if body, err := json.Marshal(c.debugMap); err == nil {
			c.acr.FireRPCEventToEngine("*.broadcast.message."+c.GetRouteId(), rpc.PublishingOption{
				Body: body,
			})
		} else {
			logger.Error("Call %s log marshal json message error: %s", c.Uuid, err.Error())
		}
	}
}

func (c *Call) OnDisconnectTrigger() {
	c.breakCall = false
	c.Iterator = c.OnDisconnectIterator
	routeIteratorOnDisconnect(c)
}

func routeIteratorOnDisconnect(call *Call) {
	for {
		if call.GetBreak() {
			logger.Debug("Call %s break disconnect route", call.GetUuid())
			break
		}
		v := call.OnDisconnectIterator.NextApp()
		if v == nil {
			break
		}

		if fn, ok := applications[v.GetName()]; ok {
			fn(call, v.GetArgs())
			continue
		}

		if v.IsBreak() {
			logger.Debug("Call %s break route", call.GetUuid())
			break
		}

		v.Execute(call.OnDisconnectIterator)
	}
}

//TODO
func routeCallIterator(call *Call, iter *router.Iterator) {
	for {
		if call.Conn.GetDisconnected() {
			//TODO setRoute onDisconnect
			logger.Warning("Call %s disconnected", call.GetUuid())
			break
		}

		if call.GetBreak() {
			logger.Debug("Call %s break route", call.GetUuid())
			break
		}
		v := iter.NextApp()
		if v == nil {
			break
		}
		//fmt.Println(v.GetName())
		if fn, ok := applications[v.GetName()]; ok {
			if fn(call, v.GetArgs()) != nil {
				logger.Debug("Call %s stop connection", call.GetUuid())
				return
			}

			if v.IsBreak() {
				call.SetBreak()
				break
			}

			continue
		}

		v.Execute(iter)
	}
}
