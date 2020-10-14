package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const tagsFile = "tags.txt"
const termsFile = "TERMS_OF_USE.txt"

func loadContentFromFile(fileName string) []string {
	f, err := os.Open(termsFile)
	if err != nil {
		log.Fatalf("error opening content file (%s):%q", fileName, err)
	}
	defer f.Close()
	r, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("error reading content from %s file:%q", fileName, err)
	}
	return strings.Split(string(r), "\n")
}

func loadTags() []string {
	return loadContentFromFile(tagsFile)
}

func loadTerms() []string {
	return loadContentFromFile(termsFile)
}
