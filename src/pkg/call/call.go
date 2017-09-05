/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
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

type Applications map[string]func(c *Call, args interface{}) error

type IBridge interface {
	GetGlobalVar(call *Call, varName string) (val string, ok bool)
	CheckBlackList(domainName, name, number string) (error, int)
	GetEmailConfig(domainName string, dataStructure interface{}) error
	GetCalendar(name, domainName string, dataStructure interface{}) error
	FindLocation(sysLength int, numbers []string, dataStructure interface{}) error
	GetDomainVariables(domainName string, dataStructure interface{}) error
	SetDomainVariable(domainName, key, value string) error
	GetRPCCommandsQueueName() string
	AddRPCCommands(uuid string) rpc.ApiT
	FireRPCEvent(body []byte, rk string) error
}

var applications Applications

type Call struct {
	Uuid                 string
	SwitchId             string
	Domain               string
	Timezone             string
	CurrentQueue         string
	DestinationNumber    string
	Iterator             *router.Iterator
	OnDisconnectIterator *router.Iterator
	LocalVariables       map[string]string
	RegExp               map[string][]string
	Conn                 *esl.SConn
	breakCall            bool
	acr                  IBridge
}

type domainVariablesT struct {
	Variables map[string]string `bson:"variables"`
}

func init() {
	regCompileVar = regexp.MustCompile(`\$\{([\s\S]*?)\}`)
	regCompileGlobalVar = regexp.MustCompile(`\$\$\{([\s\S]*?)\}`)
	regCompileLocalRegs = regexp.MustCompile(`&reg(\d+)\.\$(\d+)`)

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

	}

	var tmp string
	var i = 0
	for tmp, _ = range applications {
		i++
		logger.Info("Register application %v (%v)", tmp, i)
	}
}

func (c *Call) AddRegExp(data []string) {
	c.RegExp["reg_"+string(len(c.RegExp))] = data
}

func (c *Call) SetBreak() {
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

func (c *Call) GetGlobalVar(name string) (val string) {
	val, _ = c.acr.GetGlobalVar(c, name)
	return val
}

func (c *Call) SndMsg(app string, args string, look bool, dump bool) (esl.Event, error) {
	//TODO
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Call recovered in %v", r)
		}
	}()
	args = c.ParseString(args)
	logger.Notice("Execute %s -> %s", app, args)
	return c.Conn.SndMsg(app, args, look, dump)
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
		if strings.HasPrefix(varName, "${say_string ") || strings.HasPrefix(varName, "${hash(") { //TODO
			return varName
		}
		return c.GetChannelVar(regCompileVar.FindStringSubmatch(varName)[1])
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

	return a
}

func (c *Call) GetUuid() string {
	return c.Uuid
}

func (c *Call) ValidateApp(name string) (ok bool) {
	_, ok = applications[name]
	return
}

func MakeCall(destinationNumber string, c *esl.SConn, cf *router.CallFlow, acr IBridge) *Call {

	call := &Call{
		Timezone:          cf.Timezone,
		Uuid:              c.Uuid,
		Domain:            cf.Domain,
		Conn:              c,
		acr:               acr,
		LocalVariables:    make(map[string]string),
		SwitchId:          c.ChannelData.Header.Get("Core-UUID"),
		DestinationNumber: destinationNumber,
		RegExp:            setupNumber(cf.Number, destinationNumber),
	}

	call.Iterator = router.NewIterator(cf.Callflow, call)
	if len(cf.OnDisconnect) > 0 {
		call.OnDisconnectIterator = router.NewIterator(cf.OnDisconnect, call)
	} else {
		call.OnDisconnectIterator = nil
	}

	go setupDomainVariables(call)

	return call
}

func setupDomainVariables(call *Call) {
	var err error
	var dVars domainVariablesT

	if err = call.acr.GetDomainVariables(call.Domain, &dVars); err != nil {
		logger.Error("Call %s set domain variables db error: %s", call.Uuid, err.Error())
	}

	if len(dVars.Variables) > 0 {
		var dVarArr []interface{}
		for k, v := range dVars.Variables {
			dVarArr = append(dVarArr, k+"="+v)
		}
		err = SetVar(call, dVarArr)
		if err != nil {
			logger.Error("Call %s set domain variables error: %s", call.Uuid, err.Error())
		}
	}

	go routeIterator(call)
}

func setupNumber(reg, dest string) map[string][]string {
	storage := make(map[string][]string)
	if reg != "" {
		re := regexp.MustCompile(reg)
		d := re.FindStringSubmatch(dest)
		storage["0"] = d
	} else {
		storage["0"] = []string{}
	}
	return storage
}

func routeIterator(call *Call) {
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
			if fn(call, v.GetArgs()) != nil {
				logger.Debug("Call %s stop connection", call.GetUuid())
				break
			}
			continue
		}
		v.Execute(call.Iterator)
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
		v.Execute(call.OnDisconnectIterator)
	}
}

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
			continue
		}
		v.Execute(iter)
	}
}
