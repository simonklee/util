package log

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"github.com/simonz05/util/raven"
)

const (
	LevelFatal Level = iota
	LevelError
	LevelInfo
)

var (
	// Severity stores the log level
	Severity Level
	std      logger
	filename *string
	ravenDSN *string
)

func init() {
	flag.Var(&Severity, "log", "log level")
	filename = flag.String("log-file", "", "If non-empty, write log to this file")
	ravenDSN = flag.String("log-raven-dsn", "", "If non-empty, write to raven dsn")
	std = new(multiLogger)
}

type logger interface {
	Output(calldepth int, s string, sev Level) error
}

type multiLogger struct {
	loggers []logger
}

func (l multiLogger) Output(calldepth int, s string, sev Level) (err error) {
	if len(l.loggers) == 0 {
		l.init()
	}

	for _, w := range l.loggers {
		err = w.Output(calldepth, s, sev)

		if err != nil {
			return
		}
	}

	return
}

func (l *multiLogger) init() {
	l.loggers = append(l.loggers, &consoleLogger{sev: Severity})

	if *filename != "" {
		l.loggers = append(l.loggers, &fileLogger{fname: *filename, sev: Severity})
	}

	if *ravenDSN != "" {
		l.loggers = append(l.loggers, &ravenLogger{dsn: *ravenDSN, sev: LevelError})
	}
}

type consoleLogger struct {
	l   *log.Logger
	sev Level
}

func (l *consoleLogger) Output(calldepth int, s string, sev Level) error {
	if l.sev < sev {
		return nil
	}

	if l.l == nil {
		l.l = log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)
	}

	return l.l.Output(calldepth, s)
}

type fileLogger struct {
	l     *log.Logger
	sev   Level
	fname string
}

func (l *fileLogger) Output(calldepth int, s string, sev Level) error {
	if l.sev < sev {
		return nil
	}

	if l.l == nil {
		l.init()
	}

	return l.l.Output(calldepth, s)
}

func (l *fileLogger) init() {
	f, err := os.OpenFile(filepath.Clean(l.fname), os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}

	l.l = log.New(f, "", log.Ldate|log.Lmicroseconds)
}

type ravenLogger struct {
	l   *log.Logger
	sev Level
	dsn string
}

func (l *ravenLogger) Output(calldepth int, s string, sev Level) error {
	if l.sev < sev {
		return nil
	}

	if l.l == nil {
		l.init()
	}

	return l.l.Output(calldepth, s)
}

func (l *ravenLogger) init() {
	r, err := raven.NewClient(l.dsn, "")

	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}

	l.l = log.New(&ravenWriter{c: r}, "", log.Lshortfile)
}

type ravenWriter struct {
	c *raven.Client
}

func (w *ravenWriter) Write(p []byte) (int, error) {
	return len(p), w.c.Error(string(p))
}

func createWriter() io.Writer {
	var writers []io.Writer

	return io.MultiWriter(writers...)
}

// Level is treated as a sync/atomic int32.

// Level specifies a level of verbosity for V logs. *Level implements
// flag.Value; the -v flag is of type Level and should be modified
// only through the flag.Value interface.
type Level int32

// get returns the value of the Level.
func (l *Level) get() Level {
	return Level(atomic.LoadInt32((*int32)(l)))
}

// set sets the value of the Level.
func (l *Level) set(val Level) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

// String is part of the flag.Value interface.
func (l *Level) String() string {
	return strconv.FormatInt(int64(*l), 10)
}

// Get is part of the flag.Value interface.
func (l *Level) Get() interface{} {
	return *l
}

// Set is part of the flag.Value interface.
func (l *Level) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	l.set(Level(v))
	return nil
}

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) {
	if Severity >= LevelInfo {
		std.Output(5, fmt.Sprint(v...), LevelInfo)
	}
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	if Severity >= LevelInfo {
		std.Output(5, fmt.Sprintf(format, v...), LevelInfo)
	}
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	if Severity >= LevelInfo {
		std.Output(5, fmt.Sprintln(v...), LevelInfo)
	}
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Error(v ...interface{}) {
	if Severity >= LevelError {
		std.Output(5, fmt.Sprint(v...), LevelError)
	}
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Errorf(format string, v ...interface{}) {
	if Severity >= LevelError {
		std.Output(5, fmt.Sprintf(format, v...), LevelError)
	}
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Errorln(v ...interface{}) {
	if Severity >= LevelError {
		std.Output(5, fmt.Sprintln(v...), LevelError)
	}
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	if Severity >= LevelFatal {
		std.Output(5, fmt.Sprint(v...), LevelFatal)
		os.Exit(1)
	}
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	if Severity >= LevelFatal {
		std.Output(5, fmt.Sprintf(format, v...), LevelFatal)
		os.Exit(1)
	}
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	if Severity >= LevelFatal {
		std.Output(5, fmt.Sprintln(v...), LevelFatal)
		os.Exit(1)
	}
}
