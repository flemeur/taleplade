package errors

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Error defines a standard application error.
type Error struct {
	// Optional nested error.
	Err error

	// Logical operation.
	Op Op

	// Human-readable message.
	Message string

	// Machine-readable error type/kind.
	Kind Kind
}

// Error returns the string representation of the error message.
func (e *Error) Error() string {
	var buf bytes.Buffer

	// Print the current operation in our stack, if any.
	if e.Op != "" {
		fmt.Fprintf(&buf, "%s: ", e.Op)
	}

	// If wrapping an error, print its Error() message.
	// Otherwise print the error code & message.
	if e.Err != nil {
		buf.WriteString(e.Err.Error())
	} else {
		if e.Kind != Undefined && e.Kind.IsValid() {
			fmt.Fprintf(&buf, "<%s> ", e.Kind.String())
		}

		buf.WriteString(e.Message)
	}

	return buf.String()
}

func (e *Error) Unwrap() error { return e.Err }

// E constructs a new error from the given arguments.
func E(args ...any) error {
	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}

	e := &Error{}

	for _, arg := range args {
		switch arg := arg.(type) {
		// If any argument is nil we interpred it as being a nil error and just return nil.
		case nil:
			return nil

		case Kind:
			if !arg.IsValid() {
				panic(fmt.Sprintf("invalid error Kind provided: %d (%v)", arg, arg))
			}

			e.Kind = arg

		case string:
			e.Message = arg

		case Op:
			e.Op = arg

		case *Error:
			// Make a copy
			eCopy := *arg
			e.Err = &eCopy

		case error:
			e.Err = arg

		default:
			panic(fmt.Sprintf("invalid argument to errors.E of type %T, value %v", arg, arg))
		}
	}

	// Enforce that Err cannot coexist with message and/or Kind.
	if (e.Message != "" || e.Kind != Undefined) && e.Err != nil {
		panic("message or kind must not be provided when an error is also provided")
	}

	// Generate Op automatically if Err is provided and no Op has been set.
	if e.Err != nil && e.Op == "" {
		pc, _, _, _ := runtime.Caller(1)
		funcName := runtime.FuncForPC(pc).Name()

		lastSlash := strings.LastIndexByte(funcName, '/')
		if lastSlash < 0 {
			lastSlash = 0
		}

		e.Op = Op(strings.TrimPrefix(funcName[lastSlash:], "/"))
	}

	return e
}

type Op string

type Kind uint8

const (
	Undefined      Kind = iota // Undefined error.
	Internal                   // Internal error.
	Authentication             // Authentication required.
	Permission                 // Permission denied.
	Invalid                    // Invalid operation.
	Exists                     // Item already exists.
	NotExists                  // Item does not exist.
	Validation                 // Invalid data.
	Transient                  // Transient error, i.e. overloaded service or downtime.
	External                   // External error, i.e. service not acting as expected.
	Timeout                    // Timeout error, i.e. service that fails to respond in time.

	// Must be last. Used to check if an error's Kind is in the valid range.
	numKind
)

func (k Kind) String() string {
	switch k {
	case Internal:
		return "internal error"
	case Authentication:
		return "authentication required"
	case Permission:
		return "permission denined"
	case Invalid:
		return "invalid operation"
	case Exists:
		return "item already exists"
	case NotExists:
		return "item does not exist"
	case Validation:
		return "invalid data"
	case Transient:
		return "transient error"
	case External:
		return "external error"
	case Timeout:
		return "timeout error"

	// Required by linter
	case Undefined:
	case numKind:
	}

	return "unknown error"
}

func (k Kind) IsValid() bool {
	return k < numKind
}

// ErrorMessage returns the human-readable message of the error, if available.
// Otherwise returns a generic error message.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	var e *Error
	ok := errors.As(err, &e)

	if ok && e.Message != "" {
		return e.Message
	} else if ok && e.Err != nil {
		return ErrorMessage(e.Err)
	}

	return "An internal error has occurred"
}

// ErrorKind returns the Kind of the root error, if available. Otherwise returns Internal.
func ErrorKind(err error) Kind {
	if err == nil {
		return Undefined
	}

	var e *Error
	ok := errors.As(err, &e)

	if ok && e.Kind != Undefined && e.Kind.IsValid() {
		return e.Kind
	} else if ok && e.Err != nil {
		return ErrorKind(e.Err)
	}

	return Internal
}

// IsKind reports whether err is an *Error of the given kind.
// If err is nil then IsKind returns false.
func IsKind(kind Kind, err error) bool {
	if err == nil {
		return false
	}

	return ErrorKind(err) == kind
}

// The following functions are simply wrappers/aliases of common functions from the stdlib errors package.
// Having these functions exposed here avoids the issue of package name collision.

func New(text string) error {
	return errors.New(text)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}
