package errors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

func New(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(err error, code, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{Code: code, Message: message, Err: err}
}

func IsCode(err error, code string) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code == code
	}
	return false
}

func Is(err error) bool {
	var ae *AppError
	return errors.As(err, &ae)
}

func As(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}
