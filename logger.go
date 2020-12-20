package esl

import (
	"log"
	"os"
)

type Logger interface {
	Printf(format string, args ...interface{})
}

var logger Logger = log.New(os.Stdout, "", log.LstdFlags)

func SetLogger(l Logger) {
	logger = l
}
