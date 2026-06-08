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
	CodeNotFound = "not_found"
	// CodeUsage is for CLI-grammar failures — the invocation itself is
	// malformed: unknown command, unknown flag, missing/extra positional,
	// mutually-exclusive flags. CodeValidationError is for content failures —
	// the invocation parsed fine but a value is semantically invalid (empty
	// body, unparseable date). The split mirrors ait, which pairs CodeUsage
	// with exit 64 (see ExitCodeFor).
	CodeUsage                = "usage"
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

// exitError signals a specific shell exit code from a command handler.
// When the embedded err is nil the dispatcher skips the JSON error envelope
// — used for clean non-zero exits like 'self-update --check' reporting that
// a newer version is available without that being an error condition.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return fmt.Sprintf("exit %d", e.code)
}

func (e *exitError) Unwrap() error { return e.err }

// ExitWithCode wraps err so the binary exits with the given code. A nil err
// produces a silent non-zero exit (no stderr envelope written).
func ExitWithCode(code int, err error) error {
	return &exitError{code: code, err: err}
}

// ExitCode reports whether err carries a specific exit code. Used by main
// to translate command results into shell exit status.
func ExitCode(err error) (int, bool) {
	var e *exitError
	if errors.As(err, &e) {
		return e.code, true
	}
	return 0, false
}

// ExitCodeFor maps an error to a shell exit code from its CodedError code.
// Usage-class failures (CodeUsage) exit 64 (EX_USAGE), matching ait so a
// wrapper reads the same signal from both tools. Everything else is 1.
func ExitCodeFor(err error) int {
	var ce *CodedError
	if errors.As(err, &ce) && ce.Code == CodeUsage {
		return 64
	}
	return 1
}

// silentExit reports whether err is an exitError whose embedded cause is
// nil — Dispatch uses this to suppress WriteError for clean signal-only
// exits.
func silentExit(err error) bool {
	var e *exitError
	if errors.As(err, &e) {
		return e.err == nil
	}
	return false
}
