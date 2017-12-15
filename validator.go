package linker

import (
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"sync"
)

var Validator StructValidator = &defaultValidator{}

type (
	StructValidator interface {
		// ValidateStruct can receive any kind of type and it should never panic, even if the configuration is not right.
		// If the received type is not a struct, any validation should be skipped and nil must be returned.
		// If the received type is a struct or pointer to a struct, the validation should be performed.
		// If the struct is not valid or the validation itself fails, a descriptive error should be returned.
		// Otherwise nil must be returned.
		ValidateStruct(interface{}) error

		// RegisterValidation adds a validation Func to a Validate's map of validators denoted by the key
		// NOTE: if the key already exists, the previous validation function will be replaced.
		// NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation
		RegisterValidation(string, validator.Func) error
	}

	defaultValidator struct {
		once     sync.Once
		validate *validator.Validate
	}
)

func (v *defaultValidator) ValidateStruct(obj interface{}) error {

	if kindOfData(obj) == reflect.Struct {

		v.lazyInit()

		if err := v.validate.Struct(obj); err != nil {
			return error(err)
		}
	}

	return nil
}

func (v *defaultValidator) RegisterValidation(key string, fn validator.Func) error {
	v.lazyInit()
	return v.validate.RegisterValidation(key, fn)
}

func (v *defaultValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// add any custom validations etc. here
	})
}

func kindOfData(data interface{}) reflect.Kind {

	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}

	return Validator.ValidateStruct(obj)
}
