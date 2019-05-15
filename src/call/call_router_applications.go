package call

type ApplicationHandler func(c *Call, args interface{}) error
type Applications map[string]Application
type Application struct {
	allowNoConnect bool
	handler        ApplicationHandler
}

func (a Applications) Get(name string) (Application, bool) {
	h, ok := a[name]
	return h, ok
}

func (a Applications) Exists(name string) bool {
	_, ok := a[name]
	return ok
}

func (app Application) Execute(call *Call, args interface{}) error {
	return app.handler(call, args)
}

func (router *CallRouterImpl) initApplications() {
	router.applications = Applications{
		"setVar":        Application{true, SetVar},
		"bridge":        Application{true, Bridge},
		"answer":        Application{false, Answer},
		"hangup":        Application{false, Hangup},
		"sleep":         Application{false, Sleep},
		"agent":         Application{false, Agent},
		"amd":           Application{false, AMD},
		"blackList":     Application{false, BlackList},
		"break":         Application{true, Break},
		"cache":         Application{true, Cache},
		"log":           Application{true, Log},
		"calendar":      Application{true, Calendar},
		"callback":      Application{true, CallbackQueue},
		"cdr":           Application{true, CDR},
		"conference":    Application{false, Conference},
		"eavesdrop":     Application{false, Eavesdrop},
		"echo":          Application{false, Echo},
		"execute":       Application{true, ExecuteFunction},
		"exists":        Application{true, Exists},
		"exportVars":    Application{false, ExportVars},
		"findUser":      Application{true, FindUser},
		"flushDTMF":     Application{false, FlushDTMF},
		"geoLocation":   Application{true, GeoLocation},
		"goto":          Application{true, GoTo},
		"httpRequest":   Application{true, HttpRequest},
		"inBandDTMF":    Application{false, InBandDTMF},
		"ivr":           Application{true, IVR},
		"js":            Application{true, JavaScript},
		"limit":         Application{false, Limit},
		"math":          Application{true, Math},
		"member":        Application{true, Member},
		"originate":     Application{false, Originate},
		"park":          Application{false, Park},
		"pickup":        Application{false, Pickup},
		"playback":      Application{false, Playback},
		"queue":         Application{false, Queue},
		"ccPosition":    Application{false, QueuePosition},
		"queueStatus":   Application{true, QueueStatus},
		"receiveFax":    Application{false, ReceiveFax},
		"recordFile":    Application{false, RecordFile},
		"recordSession": Application{false, RecordSession},
		"ringback":      Application{false, RingBack},
		"schedule":      Application{false, Schedule},
		"script":        Application{false, Script},
		"sendEmail":     Application{true, SendEmail},
		"sendEvent":     Application{false, SendEvent},
		"sendSMS":       Application{true, SendSMS},
		"setArray":      Application{true, SetArray},
		"setSounds":     Application{false, SetSounds},
		"setUser":       Application{false, SetUser},
		"sipRedirect":   Application{false, SipRedirect},
		//"stream":        Application{false, Stream},
		"string":    Application{true, String},
		"stt":       Application{false, STT},
		"tts":       Application{false, TTS},
		"unSet":     Application{false, UnSet},
		"userData":  Application{true, UserData},
		"voicemail": Application{false, VoiceMail},
		//"mutex":     Application{false, Mutex},
	}

}

func (router *CallRouterImpl) GetApplication(name string) (Application, bool) {
	return router.applications.Get(name)
}

func (router *CallRouterImpl) ExistsApplication(name string) bool {
	return router.applications.Exists(name)
}
