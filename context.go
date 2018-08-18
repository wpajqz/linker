package linker

var _ Context = new(ContextTcp)
var _ Context = new(ContextUdp)
var _ Context = new(ContextWebsocket)

type (
	Context interface {
		WithValue(key interface{}, value interface{}) Context
		Value(key interface{}) interface{}
		ParseParam(data interface{}) error
		Success(body interface{})
		Error(code int, message string)
		Write(operator string, body interface{}) (int, error)
		SetRequestProperty(key, value string)
		GetRequestProperty(key string) string
		SetResponseProperty(key, value string)
		GetResponseProperty(key string) string
		LocalAddr() string
		RemoteAddr() string
	}
)
