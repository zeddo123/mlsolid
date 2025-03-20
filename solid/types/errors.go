package types

import (
	"errors"
	"fmt"
)

var (
	ErrAlreadyInUse   = errors.New("already in use")
	ErrInvalidInput   = errors.New("invalid input")
	ErrInternal       = errors.New("internal error")
	ErrBadRequest     = errors.New("bad request")
	ErrNotFound       = errors.New("not found")
	ErrNotInitialized = errors.New("not initialized")
)

func NewAlreadyInUseErr(s string) error {
	return newErr(ErrAlreadyInUse, s)
}

func NewBadRequest(s string) error {
	return newErr(ErrBadRequest, s)
}

func NewInternalErr(s string) error {
	return newErr(ErrInternal, s)
}

func NewInvalidInputErr(s string) error {
	return newErr(ErrInvalidInput, s)
}

func NewNotFoundErr(s string) error {
	return newErr(ErrNotFound, s)
}

func newErr(wrapper error, msg string) error {
	return fmt.Errorf("%w: %s", wrapper, msg)
}
