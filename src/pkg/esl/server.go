package esl


type Connection struct {
	Uuid string
	SwitchUuid string
	Disconnected bool
	ChannelData Event
}

type Server struct {

} 

type onConnect func(c *Connection)
type onDisconnect func(c *Connection)

type Event struct {
	Header Header
}

type Header map[string] []string

type Message struct {
	Header Header
}

func (e *Event) Get(string) string  {
	return ""
}
func (e *Header) Get(string) string  {
	return ""
}

func (e *Header) Add(name string, value string) string  {
	return ""
}

func (e *Header) Del(string) string  {
	return ""
}
func (e *Header) Exists(string) bool  {
	return false
}

func (c *Connection) Api(command string, args ...string) ([]byte, error) {
	return []byte{}, nil
}

func (c *Connection) BgApi(command string, args ...string) (string, error) {
	return "", nil
}

func (c *Connection) Close() {

}

func (c *Connection) GetContextName() string {
	return ""
}

func (c *Connection) Hangup(cause string) {

}


func (c *Connection) OnAnswer() bool {
	return false
}

func (c *Connection) FireEvent(name string, e *Message) ([]byte, error) {
	return []byte{}, nil
}

func (c *Connection) GetDisconnected() bool {
	return true
}

func (c Connection) SndMsg(command string, args string, lock bool, dump bool) (Event, error) {
	return Event{}, nil
}

func (s Server) Listen() error  {
	return nil
}

func NewServer(addr string, onConnect onConnect, onDisconnect onDisconnect) *Server {
	return &Server{}
}