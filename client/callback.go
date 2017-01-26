package client

type RequestStatusCallback struct {
	OnSuccess  func(ctx *Context)
	OnProgress func(progress int, status string)
	OnError    func(code int, message string)
}
