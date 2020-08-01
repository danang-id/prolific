// +build !DEBUG

package debug

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var now = time.Now().Unix()
var logDirPath = filepath.Join("logs")
var logFilePath = filepath.Join(logDirPath, fmt.Sprintf("access-log-%d.log", now))

func openLogFile() *os.File {
	if _, err := os.Stat(logDirPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(logDirPath, os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return logFile
}

func closeLogFile(logFile *os.File) {
	if err := logFile.Close(); err != nil {
		panic(err)
	}
}

func Print(v ...interface{}) {
	logFile := openLogFile()
	log.SetOutput(logFile)
	defer closeLogFile(logFile)
	log.Print(v...)
	logFile.Sync()
}

func Printf(format string, v ...interface{}) {
	logFile := openLogFile()
	log.SetOutput(logFile)
	defer closeLogFile(logFile)
	log.Printf(format, v...)
	logFile.Sync()
}

func Println(v ...interface{}) {
	logFile := openLogFile()
	log.SetOutput(logFile)
	defer closeLogFile(logFile)
	log.Println(v...)
	logFile.Sync()
}

func Fatal(v ...interface{}) {
	logFile := openLogFile()
	log.SetOutput(logFile)
	defer closeLogFile(logFile)
	log.Fatal(v...)
	logFile.Sync()
}

func Fatalf(format string, v ...interface{}) {
	logFile := openLogFile()
	log.SetOutput(logFile)
	defer closeLogFile(logFile)
	log.Fatalf(format, v...)
	logFile.Sync()
}

func Fatalln(v ...interface{}) {
	logFile := openLogFile()
	log.SetOutput(logFile)
	defer closeLogFile(logFile)
	log.Fatalln(v...)
	logFile.Sync()
}
