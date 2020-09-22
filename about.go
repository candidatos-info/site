package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	aboutFileName = "about.txt"
)

func getAboutText() string {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get current directory on loading about file, erro %v", err)
		return ""
	}
	aboutFilePath := fmt.Sprintf("%s/files/%s", currentDir, aboutFileName)
	b, err := ioutil.ReadFile(aboutFilePath)
	if err != nil {
		log.Printf("failed to open about file at path [%s], error %v", aboutFilePath, err)
		return ""
	}
	log.Println("successfully loaded about text from file")
	return string(b)
}
