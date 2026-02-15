package zap

type Field struct{}

type Logger struct{}

func (l *Logger) Debug(msg string, fields ...Field) {}
func (l *Logger) Info(msg string, fields ...Field)  {}
func (l *Logger) Warn(msg string, fields ...Field)  {}
func (l *Logger) Error(msg string, fields ...Field) {}
func (l *Logger) DPanic(msg string, fields ...Field) {}
func (l *Logger) Panic(msg string, fields ...Field) {}
func (l *Logger) Fatal(msg string, fields ...Field) {}

type SugaredLogger struct{}

func (s *SugaredLogger) Debugw(msg string, keysAndValues ...interface{}) {}
func (s *SugaredLogger) Infow(msg string, keysAndValues ...interface{})  {}
func (s *SugaredLogger) Warnw(msg string, keysAndValues ...interface{})  {}
func (s *SugaredLogger) Errorw(msg string, keysAndValues ...interface{}) {}

func String(key string, val string) Field { return Field{} }
