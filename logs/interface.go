package logs

type Args map[string]any

type Logger interface {
	Child(name string) Logger
	Info(msg string, args Args)
	Debug(msg string, args Args)
	Warn(msg string, args Args)
	Error(msg string, args Args)
}
