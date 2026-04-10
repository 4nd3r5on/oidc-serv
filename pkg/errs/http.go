package errs

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"maps"
	"net/http"
)

type respDataError struct {
	err      error
	respData map[string]any
}

// NewRespData attaches extra JSON fields to include in the HTTP response body
// alongside the "error" field. If err already carries a [respDataError], the
// existing fields are merged into the new wrapper with the new fields taking
// priority on key conflicts.
//
// Each call is immutable — a fresh map is always allocated, so the original
// error and any previously attached data are never mutated.
//
//	err = errs.NewRespData(err, map[string]any{"field": "username"})
func NewRespData(err error, respData map[string]any) error {
	merged := maps.Clone(respData)
	if rdErr, ok := errors.AsType[*respDataError](err); ok {
		maps.Copy(merged, rdErr.respData)
	}
	return &respDataError{
		err:      err,
		respData: merged,
	}
}

func (e *respDataError) Error() string {
	return e.err.Error()
}

func (e *respDataError) Unwrap() error {
	return e.err
}

// GetHTTPCode maps err to an HTTP status code based on the sentinel it carries.
// Falls back to 500 if no known sentinel is found.
func GetHTTPCode(err error) int {
	switch {
	case errors.Is(err, ErrNotImplemented):
		return http.StatusNotImplemented
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.Is(err, ErrRemoteServiceErr):
		return http.StatusBadGateway
	case errors.Is(err, ErrRateLimited):
		return http.StatusTooManyRequests
	case IsAny(err,
		ErrInvalidArgument,
		ErrMissingArgument,
		ErrOutOfRange,
	):
		return http.StatusBadRequest
	case errors.Is(err, ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case IsAny(err, ErrExists, ErrOutdated):
		return http.StatusConflict
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// IsSafeCode reports whether status is below 500, meaning it is safe to
// surface the raw error message to the client.
func IsSafeCode(status int) bool {
	return status < 500
}

// GetHTTPLogLevel returns the slog level appropriate for the given status code.
// 5xx → Error, 401/403 → Warn, everything else → Debug.
func GetHTTPLogLevel(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return slog.LevelWarn
	default:
		return slog.LevelDebug
	}
}

// GetHTTPMessageAndStatus resolves the HTTP status code and the message to
// send to the client. Message precedence:
//
//  1. Safe message attached by [SafeMsg], if present.
//  2. Generic http.StatusText for 5xx errors (prevents leaking internals).
//  3. err.Error() for all other status codes.
func GetHTTPMessageAndStatus(err error) (status int, msg string) {
	statusCode := GetHTTPCode(err)
	if !IsSafeCode(statusCode) {
		return statusCode, http.StatusText(statusCode)
	}
	return statusCode, err.Error()
}

// GetHTTPMessage is a convenience wrapper around [GetHTTPMessageAndStatus]
// that returns only the message.
func GetHTTPMessage(err error) string {
	_, msg := GetHTTPMessageAndStatus(err)
	return msg
}

// GetHTTPRespData resolves the HTTP status and builds the JSON response map.
// The map always contains an "error" key. If err carries extra fields via
// [NewRespData], they are merged in with lower priority than the "error" key.
func GetHTTPRespData(err error) (status int, response map[string]any) {
	status, msg := GetHTTPMessageAndStatus(err)
	if rdErr, ok := errors.AsType[*respDataError](err); ok {
		respData := map[string]any{"error": msg}
		maps.Copy(respData, rdErr.respData)
		return status, respData
	}
	return status, map[string]any{"error": msg}
}

// HandleHTTP is a convenience function that resolves the error into an HTTP
// response and writes it. It:
//   - resolves the status code and JSON body via [GetHTTPRespData]
//   - sets Content-Type: application/json and writes the response
func HandleHTTP(w http.ResponseWriter, err error) {
	status, resp := GetHTTPRespData(err)
	respBytes, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(respBytes)
}
