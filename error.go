package mirango

import (
	"fmt"
)

type Error struct {
	Status  int
	Code    int
	Type    string
	Message string
	Values  map[interface{}]interface{}
}

func NewError(code int, typ string, message string) *Error {
	return &Error{Code: code, Type: typ, Message: message}
}

func (s *Error) Error() string {
	return fmt.Sprintf("[%sError:%v] %v", s.Type, s.Code, s.Message)
}
