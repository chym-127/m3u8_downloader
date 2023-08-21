package utils

import (
	"fmt"
	"io"
	"log"
	"os"
)

func InitLogger(logFilePath string) *os.File {
	RemoveFileIfExist(logFilePath)
	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(io.MultiWriter(os.Stderr, f))
	return f
}
