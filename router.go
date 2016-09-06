package linker

type Router struct {
	Operator   string
	Handler    Handler
	Middleware []string
}
