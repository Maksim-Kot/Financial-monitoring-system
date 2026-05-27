package domainError

import "errors"

type Status = string

const (
	StatusValidation Status = "VALIDATION"
	StatusNotFound   Status = "NOT_FOUND"
	StatusConflict   Status = "CONFLICT"
	StatusForbidden  Status = "FORBIDDEN"
	StatusInternal   Status = "INTERNAL"
)

type Carrier interface {
	error
	Domain() DomainError
}

type DomainError struct {
	Status  Status
	Message string
}

func (e DomainError) Error() string {
	return e.Message
}

func (e DomainError) Domain() DomainError {
	return e
}

func New(code Status, msg string) DomainError {
	return DomainError{Status: code, Message: msg}
}

func Extract(err error) *DomainError {
	var c Carrier

	if !errors.As(err, &c) {
		return nil
	}
	de := c.Domain()
	return &de
}

func IsDomain(err error) bool {
	return Extract(err) != nil
}

func HasStatus(err error, status Status) bool {
	de := Extract(err)
	return de != nil && de.Status == status
}
