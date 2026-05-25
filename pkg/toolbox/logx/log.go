package logx

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

var logHelper *log.Helper

func SetLogger(logger log.Logger) {
	logHelper = log.NewHelper(log.With(logger, "caller", log.Caller(5)))
}

func Log(ctx context.Context, level log.Level, keyvals ...any) {
	logHelper.WithContext(ctx).Log(level, keyvals...)
}

func Debug(ctx context.Context, a ...any) {
	logHelper.WithContext(ctx).Debug(a...)
}

// Debugf logs a message at debug level.
func Debugf(ctx context.Context, format string, a ...any) {
	logHelper.WithContext(ctx).Debugf(format, a...)
}

// Debugw logs a message at debug level.
func Debugw(ctx context.Context, keyvals ...any) {
	logHelper.WithContext(ctx).Debugw(keyvals...)
}

// Info logs a message at info level.
func Info(ctx context.Context, a ...any) {
	logHelper.WithContext(ctx).Info(a...)
}

// Infof logs a message at info level.
func Infof(ctx context.Context, format string, a ...any) {
	logHelper.WithContext(ctx).Infof(format, a...)
}

// Infow logs a message at info level.
func Infow(ctx context.Context, keyvals ...any) {
	logHelper.WithContext(ctx).Infow(keyvals...)
}

// Warn logs a message at warn level.
func Warn(ctx context.Context, a ...any) {
	logHelper.WithContext(ctx).Warn(a...)
}

// Warnf logs a message at warnf level.
func Warnf(ctx context.Context, format string, a ...any) {
	logHelper.WithContext(ctx).Warnf(format, a...)
}

// Warnw logs a message at warnf level.
func Warnw(ctx context.Context, keyvals ...any) {
	logHelper.WithContext(ctx).Warnw(keyvals...)
}

// Error logs a message at error level.
func Error(ctx context.Context, a ...any) {
	logHelper.WithContext(ctx).Error(a...)
}

// Errorf logs a message at error level.
func Errorf(ctx context.Context, format string, a ...any) {
	logHelper.WithContext(ctx).Errorf(format, a...)
}

// Errorw logs a message at error level.
func Errorw(ctx context.Context, keyvals ...any) {
	logHelper.WithContext(ctx).Errorw(keyvals...)
}

// Fatal logs a message at fatal level.
func Fatal(ctx context.Context, a ...any) {
	logHelper.WithContext(ctx).Fatal(a...)
}

// Fatalf logs a message at fatal level.
func Fatalf(ctx context.Context, format string, a ...any) {
	logHelper.WithContext(ctx).Fatalf(format, a...)
}

// Fatalw logs a message at fatal level.
func Fatalw(ctx context.Context, keyvals ...any) {
	logHelper.WithContext(ctx).Fatalw(keyvals...)
}

func init() {
	SetLogger(log.GetLogger())
}
