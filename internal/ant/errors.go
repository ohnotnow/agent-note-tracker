package ant

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Error code vocabulary emitted in the {"error": {"code": ...}} envelope.
// Treat these as a stable contract: agents may branch on them, so renaming
// any value is a breaking change. Mirrors ait's vocabulary so that an agent
// hopping between the two tools can use one switch on err.code.
const (
	CodeNotFound             = "not_found"
	CodeValidationError      = "validation_error"
	CodeConflict             = "conflict"
	CodeConfirmationRequired = "confirmation_required"
	CodeUninitialised        = "uninitialised"
	CodeInternalError        = "internal_error"
)

// CodedError is an error tagged with one of the Code* constants. Command
// handlers return one of these whenever the failure mode is something an
// agent might branch on (not-found, validation, conflict, …); anything else
// falls through to "internal_error" in WriteError.
type CodedError struct {
	Code    string
	Message string
	cause   error
}

func (e *CodedError) Error() string { return e.Message }
func (e *CodedError) Unwrap() error { return e.cause }

// NewError builds a CodedError with a formatted message.
func NewError(code, format string, args ...any) *CodedError {
	return &CodedError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// errorEnvelope is the JSON shape written to stderr on command failure.
type errorEnvelope struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError serialises err as a JSON envelope to stderr. Errors without an
// explicit code default to CodeInternalError; callers that want a specific
// code should return a CodedError.
func (a *App) WriteError(err error) {
	code := CodeInternalError
	var ce *CodedError
	if errors.As(err, &ce) {
		code = ce.Code
	}
	enc := json.NewEncoder(a.Stderr)
	enc.SetIndent("", "  ")
	_ = enc.Encode(errorEnvelope{Error: errorPayload{Code: code, Message: err.Error()}})
}
