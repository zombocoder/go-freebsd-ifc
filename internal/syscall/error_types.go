//go:build freebsd
// +build freebsd

package syscall

import (
	"errors"
	"fmt"
)

// Error types for different failure modes
var (
	// ErrPermission indicates the operation requires elevated privileges
	ErrPermission = errors.New("operation not permitted (need root)")

	// ErrNotFound indicates the requested resource does not exist
	ErrNotFound = errors.New("resource not found")

	// ErrExists indicates the resource already exists
	ErrExists = errors.New("resource already exists")

	// ErrInvalidArgument indicates an invalid parameter was provided
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrBusy indicates the resource is currently in use
	ErrBusy = errors.New("resource busy")

	// ErrNotSupported indicates the operation is not supported
	ErrNotSupported = errors.New("operation not supported")

	// ErrNetworkDown indicates the network interface is down
	ErrNetworkDown = errors.New("network is down")

	// ErrAddressInUse indicates the address is already in use
	ErrAddressInUse = errors.New("address already in use")

	// ErrSyscall is a generic syscall error wrapper
	ErrSyscall = errors.New("syscall error")
)

// OperationError provides context about where an error occurred
type OperationError struct {
	Op        string // Operation name (e.g., "CreateBridge", "AddIPAddress")
	Interface string // Interface name if applicable
	Err       error  // Underlying error
}

func (e *OperationError) Error() string {
	if e.Interface != "" {
		return fmt.Sprintf("%s(%s): %v", e.Op, e.Interface, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

// NewOpError creates a new operation error with context
func NewOpError(op, iface string, err error) error {
	if err == nil {
		return nil
	}
	return &OperationError{
		Op:        op,
		Interface: iface,
		Err:       err,
	}
}

// ValidationError represents an error from input validation
type ValidationError struct {
	Field string // Field name
	Value string // Invalid value
	Msg   string // Error message
}

func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation error: %s '%s': %s", e.Field, e.Value, e.Msg)
	}
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Msg)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidArgument
}

// NewValidationError creates a new validation error
func NewValidationError(field, value, msg string) error {
	return &ValidationError{
		Field: field,
		Value: value,
		Msg:   msg,
	}
}

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsPermission checks if an error is a permission error
func IsPermission(err error) bool {
	return errors.Is(err, ErrPermission)
}

// IsExists checks if an error is an "already exists" error
func IsExists(err error) bool {
	return errors.Is(err, ErrExists)
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
