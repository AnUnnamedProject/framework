// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite
)

const (
	CRITICAL Level = iota
	ERROR
	WARNING
	INFO
	DEBUG
)

var (
	colors = []string{
		CRITICAL: colorSeq(colorMagenta),
		ERROR:    colorSeq(colorRed),
		WARNING:  colorSeq(colorYellow),
		INFO:     colorSeq(colorGreen),
		DEBUG:    colorSeq(colorCyan),
	}
	lvlNames = []string{
		"CRIT",
		"ERRO",
		"WARN",
		"INFO",
		"DEBU",
	}
	currentBackend io.Writer
)

type (
	color int

	// Level is the log gravity level
	Level int

	// Logger is the logger structure
	Logger struct {
		*log.Logger
	}
)

// Log is the framework global Logger
var Log *Logger

func colorSeq(color color) string {
	return fmt.Sprintf("\033[%dm", int(color))
}

// NewLogger create and return a new Logger instance
func NewLogger(out io.Writer) *Logger {
	currentBackend = out
	l := &Logger{log.New(out, "", log.LstdFlags)}
	return l
}

// log is the private function to
func (l *Logger) log(lvl Level, str string) {
	buf := &bytes.Buffer{}

	if currentBackend == os.Stdout {
		col := colors[lvl]
		buf.Write([]byte(col))
	}
	buf.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	buf.WriteString(" " + lvlNames[lvl] + " ")
	if currentBackend == os.Stdout {
		buf.Write([]byte("\033[0m"))
	}
	buf.WriteString(str)
	fmt.Println(buf.String())
}

// Critical is an alias to log(CRITICAL, str)
func (l *Logger) Critical(str string) {
	l.log(CRITICAL, str)
}

// Criticalf is an alias to log(CRITICAL, str)
func (l *Logger) Criticalf(format string, a ...interface{}) {
	l.log(CRITICAL, fmt.Sprintf(format, a))
}

// Error is an alias to log(ERROR, err)
func (l *Logger) Error(err error) {
	l.log(ERROR, err.Error())
}

// Errorf is an alias to log(ERROR, err)
func (l *Logger) Errorf(format string, a ...interface{}) {
	l.log(ERROR, fmt.Errorf(format, a).Error())
}

// Warning is an alias to log(WARNING, str)
func (l *Logger) Warning(str string) {
	l.log(WARNING, str)
}

// Warningf is an alias to log(WARNING, str)
func (l *Logger) Warningf(format string, a ...interface{}) {
	l.log(WARNING, fmt.Sprintf(format, a))
}

// Info is an alias to log(INFO, str)
func (l *Logger) Info(str string) {
	l.log(INFO, str)
}

// Infof is an alias to log(INFO, str)
func (l *Logger) Infof(format string, a ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, a))
}

// Debug is an alias to log(DEBUG, str)
func (l *Logger) Debug(str string) {
	l.log(DEBUG, str)
}

// Debugf is an alias to log(DEBUG, str)
func (l *Logger) Debugf(format string, a ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, a))
}
