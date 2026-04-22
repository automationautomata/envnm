package grpc

import (
	"context"
	"envmn/logs"

	grpclogs "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

type logDecorator struct {
	inner logs.Logger
}

func (log *logDecorator) Log(ctx context.Context, level grpclogs.Level, msg string, fields ...any) {
	args := toArgs(fields)
	switch level {
	case grpclogs.LevelDebug:
		log.inner.Debug(msg, args)
	case grpclogs.LevelInfo:
		log.inner.Info(msg, args)
	case grpclogs.LevelWarn:
		log.inner.Warn(msg, args)
	case grpclogs.LevelError:
		log.inner.Error(msg, args)
	}
}

func toArgs(fields ...any) logs.Args {
	argsLen := len(fields)
	if argsLen%2 != 0 {
		argsLen += 1
		fields = append(fields, nil)
	}

	args := make(logs.Args, argsLen)
	for i := 0; i < argsLen; i += 2 {
		args[fields[i]] = fields[i+1]
	}
	return args
}
