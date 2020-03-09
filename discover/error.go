package discover

import "errors"

// error
var (
	ErrorServicePathNil   = errors.New("service path nil")
	ErrorServiceTypeNil   = errors.New("service type nil")
	ErrorServiceNameNil   = errors.New("service name nil")
	ErrorServiceIPInfoNil = errors.New("service ip info nil")
)
