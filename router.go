package linker

import (
	"hash/crc32"
	"strconv"
)

type Router struct {
	handlerContainer map[uint32]Handler
	routerMiddleware map[uint32][]Middleware
	middleware       []Middleware
}

func NewRouter() *Router {
	return &Router{
		handlerContainer: make(map[uint32]Handler),
		routerMiddleware: make(map[uint32][]Middleware),
	}
}

// 注册
func (r *Router) HandleFunc(pattern string, handler Handler, middleware ...Middleware) *Router {
	operator := crc32.ChecksumIEEE([]byte(pattern))
	if operator <= OPERATOR_MAX {
		panic("Unavailable operator, the value of crc32 need less than " + strconv.Itoa(OPERATOR_MAX))
	}

	r.routerMiddleware[operator] = append(r.routerMiddleware[operator], middleware...)

	if _, ok := r.handlerContainer[operator]; !ok {
		r.handlerContainer[operator] = handler
	}

	return r
}

// 添加请求需要进行处理的中间件
func (r *Router) Use(middleware ...Middleware) *Router {
	r.middleware = append(r.middleware, middleware...)

	return r
}
