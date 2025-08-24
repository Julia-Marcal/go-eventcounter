package logger

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

var (
	InfoColor    = color.New(color.FgCyan)
	SuccessColor = color.New(color.FgGreen)
	WarningColor = color.New(color.FgYellow)
	ErrorColor   = color.New(color.FgRed)
	ProcessColor = color.New(color.FgMagenta)
	SystemColor  = color.New(color.FgBlue)
)

func Info(format string, v ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, v...)
	InfoColor.Printf("[%s] [INFO] %s\n", timestamp, message)
}

func Success(format string, v ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, v...)
	SuccessColor.Printf("[%s] [SUCESSO] %s\n", timestamp, message)
}

func Warning(format string, v ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, v...)
	WarningColor.Printf("[%s] [AVISO] %s\n", timestamp, message)
}

func Error(format string, v ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, v...)
	ErrorColor.Printf("[%s] [ERRO] %s\n", timestamp, message)
}

func Process(format string, v ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, v...)
	ProcessColor.Printf("[%s] [PROCESSO] %s\n", timestamp, message)
}

func System(format string, v ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, v...)
	SystemColor.Printf("[%s] [SISTEMA] %s\n", timestamp, message)
}

func Fatal(v ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	ErrorColor.Printf("[%s] [FATAL] ", timestamp)
	log.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, v...)
	ErrorColor.Printf("[%s] [FATAL] %s\n", timestamp, message)
	log.Fatal()
}
