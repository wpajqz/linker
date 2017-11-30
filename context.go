package linker

type Context interface {
	WithValue(key interface{}, value interface{}) Context
	Value(key interface{}) interface{}
	SetContentType(contentType string)
	ParseParam(data interface{}) error
	Success(body interface{})
	Error(code int, message string)
	Write(operator string, body interface{}) (int, error)
	WriteBinary(operator string, data []byte) (int, error)
	SetRequestProperty(key, value string)
	GetRequestProperty(key string) string
	SetResponseProperty(key, value string)
	GetResponseProperty(key string) string
}
