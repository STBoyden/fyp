package logging

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/TwiN/go-color"
)

type logLevel int

const (
	TRACE logLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	_LogLevelAmount
)

var colours = [...]string{
	color.Gray,
	color.Green,
	color.Cyan,
	color.Yellow,
	color.Red,
	color.Purple,
}

func getColourForLevel(level logLevel) string {
	if level.String() == "UNKNOWN" {
		return color.GrayBackground
	}

	return colours[level]
}

func (level logLevel) String() string {
	strings := [...]string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	if level < TRACE || level >= _LogLevelAmount {
		return "UNKNOWN"
	}

	return strings[level]
}

type Logger struct {
	inner *log.Logger
}

func NewClient() *Logger {
	return &Logger{
		log.New(os.Stdout, color.Ize(color.Black, color.Ize(color.CyanBackground, ":: CLIENT ::"))+" ", log.LstdFlags|log.LUTC),
	}
}

func NewServer() *Logger {
	return &Logger{
		log.New(os.Stdout, color.Ize(color.Black, color.Ize(color.GreenBackground, ":: SERVER ::"))+" ", log.LstdFlags|log.LUTC),
	}
}

func (l *Logger) log(output io.Writer, level logLevel, format string, v ...any) {
	colourisedLevel := color.Ize(getColourForLevel(level), level.String())
	message := fmt.Sprintf(format, v...)

	l.inner.SetOutput(output)
	l.inner.Printf("[%s]: %s", colourisedLevel, message)
}

func (l *Logger) Tracef(format string, v ...any) {
	l.log(os.Stdout, TRACE, format, v...)
}

func (l *Logger) Trace(message string) {
	l.log(os.Stdout, TRACE, message)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.log(os.Stdout, DEBUG, format, v...)
}

func (l *Logger) Debug(message string) {
	l.log(os.Stdout, DEBUG, message)
}

func (l *Logger) Infof(format string, v ...any) {
	l.log(os.Stdout, INFO, format, v...)
}

func (l *Logger) Info(message string) {
	l.log(os.Stdout, INFO, message)
}

func (l *Logger) Warnf(format string, v ...any) {
	l.log(os.Stderr, WARN, format, v...)
}

func (l *Logger) Warn(message string) {
	l.log(os.Stderr, WARN, message)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.log(os.Stderr, ERROR, format, v...)
}

func (l *Logger) Error(message string) {
	l.log(os.Stderr, ERROR, message)
}
