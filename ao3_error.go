package ao3

import "github.com/pkg/errors"

type AO3Error struct {
	code int
	err  error
}

func NewError(code int, message string) *AO3Error {
	return &AO3Error{
		code: code,
		err:  errors.New(message),
	}
}

func WrapError(code int, err error, message string) *AO3Error {
	return &AO3Error{
		code: code,
		err:  errors.Wrap(err, message),
	}
}

func (e *AO3Error) Code() int {
	return e.code
}

func (e *AO3Error) Error() string {
	return e.err.Error()
}
