package utils

import (
	"fmt"
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

	log.SetOutput(f)
	return f
}

func Println(msg ...any) {
	fmt.Println(msg...)
	log.Println(msg...)
}

func Error(msg ...any) {
	fmt.Println(msg...)
	log.Println(msg...)
	log.Fatalln("Mission failed.")
}
