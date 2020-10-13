package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const tagsFile = "tags.txt"

func mustLoadTags() []string {
	f, err := os.Open(tagsFile)
	if err != nil {
		log.Fatalf("error opening tags file (%s):%q", tagsFile, err)
	}
	defer f.Close()
	r, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("error reading tags from %s file:%q", tagsFile, err)
	}
	return strings.Split(string(r), "\n")
}
