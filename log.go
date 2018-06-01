package lg

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var std = New(os.Stderr, "", LstdFlags)

const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	Lrelativefile
	LstdFlags = Ldate | Ltime
)

const (
	Ldebug = iota
	Linfo
	Lwarn
	Lerror
	Lpanic
	Lfatal
)

var levels = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"PANIC",
	"FATAL",
}

func ParseLogLevel(levelstr string, verbose bool) (int, error) {
	lvl := Linfo

	switch strings.ToLower(levelstr) {
	case "debug":
		lvl = Ldebug
	case "info":
		lvl = Linfo
	case "warn":
		lvl = Lwarn
	case "error":
		lvl = Lerror
	case "fatal":
		lvl = Lfatal
	default:
		return lvl, fmt.Errorf("invalid log-level '%s'", levelstr)
	}
	if verbose {
		lvl = Ldebug
	}
	return lvl, nil
}

func itoa(buf *[]byte, i int, wid int) {
	var u uint = uint(i)
	if u == 0 && wid <= 1 {
		*buf = append(*buf, '0')
		return
	}

	var b [32]byte
	var bp int = len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	*buf = append(*buf, b[bp:]...)
}

type Logger struct {
	mu        sync.Mutex
	prefix    string
	flag      int
	out       io.Writer
	buf       []byte
	level     int
	calldepth int
}

func New(out io.Writer, prefix string, flag int) *Logger {
	level := Ldebug
	return &Logger{out: out, prefix: prefix, flag: flag, level: level, calldepth: 2}
}

func (l *Logger) formatHeader(buf *[]byte, t time.Time, level int, file string, line int) {
	*buf = append(*buf, l.prefix...)
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			*buf = append(*buf, '[')
			itoa(buf, year, 4)
			*buf = append(*buf, '-')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '-')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&(Lmicroseconds) != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, "] "...)
		}
	}
	*buf = append(*buf, levels[level]...)
	*buf = append(*buf, ": "...)
	if l.flag&(Lshortfile|Llongfile|Lrelativefile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		} else if l.flag&Lrelativefile != 0 {
			relative := file
			temp := 0
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					temp++
				}
				if file[i] == '/' && temp == 2 {
					relative = file[i+1:]
					break
				}
			}
			file = relative
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func (l *Logger) Output(level int, s string) error {
	if l.level > level {
		return nil
	}

	now := time.Now()
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.flag&(Lshortfile|Lrelativefile|Llongfile) != 0 {
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(l.calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}

	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, level, file, line)
	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}

func (l *Logger) Calldepth() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.calldepth
}

func (l *Logger) SetCalldepth(calldepth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.calldepth = calldepth
}

func (l *Logger) Level() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *Logger) SetLevel(level int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) SetLevelByName(name string) error {
	level, err := ParseLogLevel(name, false)
	if err != nil {
		return err
	}

	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
	return nil
}

func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = out
}

func (l *Logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

func (l *Logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

func (l *Logger) Debug(v ...interface{}) {
	l.Output(Ldebug, fmt.Sprint(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Output(Ldebug, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.Output(Linfo, fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(Linfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.Output(Lwarn, fmt.Sprint(v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Output(Lwarn, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.Output(Lerror, fmt.Sprint(v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Output(Lerror, fmt.Sprintf(format, v...))
}

func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Output(Lpanic, s)
	panic(s)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Output(Lpanic, s)
	panic(s)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Output(Lfatal, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Output(Lfatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func DefaultLogger() *Logger {
	return std
}

func Calldepth() int {
	std.mu.Lock()
	defer std.mu.Unlock()
	return std.calldepth
}

func SetCalldepth(calldepth int) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.calldepth = calldepth
}

func Level() int {
	std.mu.Lock()
	defer std.mu.Unlock()
	return std.level
}

func SetLevel(level int) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.level = level
}

func SetOutput(out io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = out
}

func Prefix() string {
	std.mu.Lock()
	defer std.mu.Unlock()
	return std.prefix
}

func SetPrefix(prefix string) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.prefix = prefix
}

func Flags() int {
	std.mu.Lock()
	defer std.mu.Unlock()
	return std.flag
}

func SetFlags(flag int) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.flag = flag
}

func Debug(v ...interface{}) {
	std.Output(Ldebug, fmt.Sprint(v...))
}

func SetLevelByName(name string) error {
	level, err := ParseLogLevel(name, false)
	if err != nil {
		return err
	}

	std.mu.Lock()
	std.level = level
	std.mu.Unlock()
	return nil
}

func Debugf(format string, v ...interface{}) {
	std.Output(Ldebug, fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	std.Output(Linfo, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	std.Output(Linfo, fmt.Sprintf(format, v...))
}

func Warn(v ...interface{}) {
	std.Output(Lwarn, fmt.Sprint(v...))
}

func Warnf(format string, v ...interface{}) {
	std.Output(Lwarn, fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	std.Output(Lerror, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	std.Output(Lerror, fmt.Sprintf(format, v...))
}

func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	std.Output(Lpanic, s)
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	std.Output(Lpanic, s)
	panic(s)
}

func Fatal(v ...interface{}) {
	std.Output(Lfatal, fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	std.Output(Lfatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}
