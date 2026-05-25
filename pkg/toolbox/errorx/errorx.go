package errorx

import (
	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/pkg/errors"
)

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string, args ...any) error {
	return errors.Errorf(message, args...)
}

// WithStack annotates err with a stack trace at the point WithStack was called.
// If err is nil, WithStack returns nil.
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	if hasStackTrace(err) || isKratosError(err) {
		return err
	}

	return errors.WithStack(err)
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string, args ...any) error {
	if err == nil {
		return nil
	}

	if hasStackTrace(err) || isKratosError(err) {
		return errors.WithMessagef(err, message, args...)
	}

	return errors.Wrapf(err, message, args...)
}

// hasStackTrace 检查错误是否包含堆栈跟踪信息
func hasStackTrace(err error) bool {
	_, ok := err.(interface {
		StackTrace() errors.StackTrace
	})

	return ok
}

// isKratosError 检查错误是否为业务错误
func isKratosError(err error) bool {
	var kErr *kratosErrors.Error

	return errors.As(err, &kErr)
}
