/*
logging provides the Logger struct that provides standard logging functions, intended
for use with the cmd/client and cmd/server packages.
*/
package logging

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/TwiN/go-color"
)

type logLevel int

// Standard log levels. The lower the level, the more information is shown.
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

/*
Logger is a struct that acts as a wrapper over the stdlib log.Logger, providing colourful
output that is separated by client and server distinctions.
*/
type Logger struct {
	inner *log.Logger
}

/*
NewClient returns a *Logger that is suited for logging with cmd/client. Text is black,
and the background is cyan. The time information uses the UTC timezone.
*/
func NewClient() *Logger {
	return &Logger{
		log.New(os.Stdout, color.Ize(color.Black, color.Ize(color.CyanBackground, ":: CLIENT ::"))+" ", log.LstdFlags|log.LUTC),
	}
}

/*
NewClient returns a *Logger that is suited for logging with cmd/server. Text is black,
and the background is green. The time information uses the UTC timezone.
*/
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

// Tracef outputs a formattable message at the TRACE level.
func (l *Logger) Tracef(format string, v ...any) {
	l.log(os.Stdout, TRACE, format, v...)
}

// Trace outputs a simple message at the TRACE level.
func (l *Logger) Trace(message string) {
	l.log(os.Stdout, TRACE, message)
}

// Debugf outputs a formattable message at the DEBUG level.
func (l *Logger) Debugf(format string, v ...any) {
	l.log(os.Stdout, DEBUG, format, v...)
}

// Debug outputs a simple message at the DEBUG level.
func (l *Logger) Debug(message string) {
	l.log(os.Stdout, DEBUG, message)
}

// Infof outputs a formattable message at the INFO level.
func (l *Logger) Infof(format string, v ...any) {
	l.log(os.Stdout, INFO, format, v...)
}

// Info outputs a simple message at the INDOlevel.
func (l *Logger) Info(message string) {
	l.log(os.Stdout, INFO, message)
}

// Warnf outputs a formattable message at the WARN level.
func (l *Logger) Warnf(format string, v ...any) {
	l.log(os.Stderr, WARN, format, v...)
}

// Warn outputs a simple message at the WARN level.
func (l *Logger) Warn(message string) {
	l.log(os.Stderr, WARN, message)
}

// Errorf outputs a formattable message at the ERROR level.
func (l *Logger) Errorf(format string, v ...any) {
	l.log(os.Stderr, ERROR, format, v...)
}

// Error outputs a simple message at the ERROR level.
func (l *Logger) Error(message string) {
	l.log(os.Stderr, ERROR, message)
}
