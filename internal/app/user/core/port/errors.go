package port

import (
	"errors"
)

var (
	// ErrNotFound error means that requested entity is not found.
	ErrNotFound = errors.New("not found")

	// ErrAttemptedSettingLocationOfNonExistentUser error means that User with UserID that
	// SetLocation was called with does not exist.
	ErrAttemptedSettingLocationOfNonExistentUser = errors.New("attempted setting location of non-existent user")

	// ErrAlreadyExists error means that the entity a client attempted to create already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidUsername error means that a client attempted to set invalid username.
	ErrInvalidUsername = errors.New("invalid username")
)

// TODO add description
type InvalidLocationErrorViolation struct {
	Subject string
	Value   float64
}

type InvalidLocationError struct {
	Violations []InvalidLocationErrorViolation
}

func (e *InvalidLocationError) Error() string {
	return "invalid location"
}

type InvalidArgumentError struct{}

func (e *InvalidArgumentError) Error() string {
	return "invalid argument"
}
