package client

type (
	ReadyStateCallback struct {
		Open  func()
		Close func()
		Error func(err string)
	}

	RequestStatusCallback struct {
		Start   func()
		End     func()
		Success func(header, body []byte)
		Error   func(code int, message string)
	}
)

func (r *ReadyStateCallback) OnOpen() {
	if r.Open != nil {
		r.Open()
	}
}

func (r *ReadyStateCallback) OnClose() {
	if r.Close != nil {
		r.Close()
	}
}

func (r *ReadyStateCallback) OnError(err string) {
	if r.Error != nil {
		r.Error(err)
	}
}

func (r RequestStatusCallback) OnStart() {
	if r.Start != nil {
		r.Start()
	}
}

func (r RequestStatusCallback) OnSuccess(header, body []byte) {
	if r.Success != nil {
		r.Success(header, body)
	}
}

func (r RequestStatusCallback) OnError(code int, message string) {
	if r.Error != nil {
		r.Error(code, message)
	}
}

func (r RequestStatusCallback) OnEnd() {
	if r.End != nil {
		r.End()
	}
}
