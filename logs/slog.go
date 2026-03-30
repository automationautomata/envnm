package logs

import (
	"log/slog"
)

type SlogAdapter struct {
	logger *slog.Logger
}

func NewSlogAdapter(l *slog.Logger) *SlogAdapter {
	return &SlogAdapter{logger: l}
}

func (s *SlogAdapter) Child(name string) Logger {
	return NewSlogAdapter(s.logger.WithGroup(name))
}

func (s *SlogAdapter) Info(msg string, args Args) {
	s.logger.Info(msg, toAttrs(args)...)
}

func (s *SlogAdapter) Debug(msg string, args Args) {
	s.logger.Debug(msg, toAttrs(args)...)
}

func (s *SlogAdapter) Warn(msg string, args Args) {
	s.logger.Warn(msg, toAttrs(args)...)
}

func (s *SlogAdapter) Error(msg string, args Args) {
	s.logger.Error(msg, toAttrs(args)...)
}

func toAttrs(args Args) []any {
	if len(args) == 0 {
		return nil
	}

	out := make([]any, 0, len(args)*2)
	for k, v := range args {
		out = append(out, k, v)
	}
	return out
}
