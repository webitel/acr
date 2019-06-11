package provider

type CallServer interface {
	Start()
	Stop()
	Consume() <-chan Connection
}

type Connection interface {
	Id() string
	NodeId() string
	Context() string
	Destination() string
	Direction() string
	Stopped() bool
	SetDirection(direction string) error
	Set(key, value string) error
	Export(data string) error
	GetGlobalVariables() (map[string]string, error)
	GetVariable(name string) (value string)
	Api(cmd string) ([]byte, error)
	BgExecute(app, args string) error
	Execute(app, args string) error
	SendEvent(m map[string]string, name string) error
	DumpVariables() map[string]string
	HangupCause() string
	WaitForDisconnect()
	Hangup(cause string) error
}
