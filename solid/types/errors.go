package types

import (
	"errors"
	"fmt"
)

var (
	ErrAlreadyInUse   = errors.New("already in use") //nolint: revive
	ErrInvalidInput   = errors.New("invalid input")
	ErrInternal       = errors.New("internal error")
	ErrBadRequest     = errors.New("bad request")
	ErrNotFound       = errors.New("not found")
	ErrNotInitialized = errors.New("not initialized")
)

// NewAlreadyInUseErr returns new ErrAlreadyInUse error.
func NewAlreadyInUseErr(s string) error {
	return newErr(ErrAlreadyInUse, s)
}

// NewBadRequest returns new ErrBadRequest error.
func NewBadRequest(s string) error {
	return newErr(ErrBadRequest, s)
}

// NewInternalErr returns new ErrInternal error.
func NewInternalErr(s string) error {
	return newErr(ErrInternal, s)
}

// NewInvalidInputErr returns new ErrInvalidInput error.
func NewInvalidInputErr(s string) error {
	return newErr(ErrInvalidInput, s)
}

// NewNotFoundErr returns new ErrNotFound error.
func NewNotFoundErr(s string) error {
	return newErr(ErrNotFound, s)
}

func newErr(wrapper error, msg string) error {
	return fmt.Errorf("%w: %s", wrapper, msg)
}
