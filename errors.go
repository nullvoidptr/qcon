package quickconnect

import "errors"

var (
	ErrTimeout           error = errors.New("operation timed out")
	ErrCancelled         error = errors.New("operation cancelled")
	ErrInvalidID         error = errors.New("invalid server ID")
	ErrCannotAccess      error = errors.New("cannot access any URLs")
	ErrParse             error = errors.New("response parse error")
	ErrPingFailure       error = errors.New("ping response failure")
	ErrUnknownCommand    error = errors.New("unknown command")
	ErrUnknownServerType error = errors.New("unknown server type")
)
