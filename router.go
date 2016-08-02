package linker

type Router struct {
	Operator   int32
	Handler    Handler
	Middleware []string
}
