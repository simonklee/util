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

type severity int32

const (
	LevelFatal Level = iota
	LevelError
	LevelInfo 
)

var (
	std      *logger
	sev      Level
	filename *string
	ravenDSN *string
)

func init() {
	flag.Var(&sev, "log", "log level")
	filename = flag.String("log-file", "", "If non-empty, write log to this file")
	ravenDSN = flag.String("log-raven-dsn", "", "If non-empty, write to raven dsn")
	std = new(logger)
}

type logger struct {
	l *log.Logger
}

func (l *logger) Output(calldepth int, s string) error {
	if l.l == nil {
		l.l = log.New(createWriter(), "", log.Ldate|log.Lmicroseconds)
	}

	return l.l.Output(calldepth, s)
}

func createWriter() io.Writer {
	var writers []io.Writer
	writers = append(writers, os.Stderr)

	if *filename != "" {
		if f, err := createFile(); err != nil {
			os.Stderr.Write([]byte(err.Error()))
			os.Exit(1)
		} else {
			writers = append(writers, f)
		}
	}

	if *ravenDSN != "" {
		if r, err := raven.NewClient(*ravenDSN, ""); err != nil {
			os.Stderr.Write([]byte(err.Error()))
			os.Exit(1)
		} else {
			writers = append(writers, &ravenWriter{c: r})
		}
	}

	return io.MultiWriter(writers...)
}

func createFile() (f *os.File, err error) {
	fname := filepath.Clean(*filename)
	return os.Create(fname)
}

type ravenWriter struct {
	c *raven.Client
}

func (w *ravenWriter) Write(p []byte) (int, error) {
	return len(p), w.c.Error(string(p))
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
	if sev >= LevelInfo {
		std.Output(2, fmt.Sprint(v...))
	}
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	if sev >= LevelInfo {
		std.Output(2, fmt.Sprintf(format, v...))
	}
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	if sev >= LevelInfo {
		std.Output(2, fmt.Sprintln(v...))
	}
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Error(v ...interface{}) {
	if sev >= LevelError {
		std.Output(2, fmt.Sprint(v...))
		os.Exit(1)
	}
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Errorf(format string, v ...interface{}) {
	if sev >= LevelError {
		std.Output(2, fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Errorln(v ...interface{}) {
	if sev >= LevelError {
		std.Output(2, fmt.Sprintln(v...))
		os.Exit(1)
	}
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	if sev >= LevelFatal {
		std.Output(2, fmt.Sprint(v...))
		os.Exit(1)
	}
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	if sev >= LevelFatal {
		std.Output(2, fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	if sev >= LevelFatal {
		std.Output(2, fmt.Sprintln(v...))
		os.Exit(1)
	}
}
