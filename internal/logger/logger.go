package logger

import (
	"bytes"
	"log"
	"os"
)

//nolint:golint,gochecknoglobals
var (
	infoLog  *log.Logger
	errorLog *log.Logger
	fatalLog *log.Logger
	debugLog *log.Logger
)

const (
	INFO  = iota
	ERROR = iota
	FATAL = iota
	DEBUG = iota
)

func InitializeLogger() {
	infoLog = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime)
	errorLog = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime)
	fatalLog = log.New(os.Stderr,
		"FATAL: ",
		log.Ldate|log.Ltime)
	debugLog = log.New(os.Stdout,
		"DEBUG: ",
		log.Ldate|log.Ltime)
}

func WriteLogMessage(level int, message ...string) {
	var buffer bytes.Buffer

	for _, m := range message {
		buffer.WriteString(m)
		buffer.WriteString(" ")
	}

	switch level {
	case INFO:
		infoLog.Println(buffer.String())
	case ERROR:
		errorLog.Println(buffer.String())
	case FATAL:
		fatalLog.Println(buffer.String())
		os.Exit(1)
	case DEBUG:
		debugLog.Println(buffer.String())
	}
}
