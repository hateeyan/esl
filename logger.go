package esl

import (
	"log"
	"os"
)

type Logger interface {
	Printf(format string, args ...interface{})
}

var defaultLogger = log.New(os.Stderr, "", log.LstdFlags)
