package logger

import (
	"log"
	"os"
	"sync"
)
var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

var once sync.Once

func initLogger() {
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func ensureInit() {
	once.Do(initLogger)
}

func Info(v ...any) {
	ensureInit()
	infoLogger.Println(v...)
}

func Error(v ...any) {
	ensureInit()
	errorLogger.Println(v...)
}