package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gocarina/gocsv"
	"golang.org/x/text/encoding/charmap"
)

const (
	partnersFileName = "partners.csv"
)

type partner struct {
	Link    string `json:"link" csv:"link"`
	IconURL string `json:"icon_url" csv:"icon_url"`
}

func getPartners() []*partner {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get current directory on loading partners file, erro %v", err)
		return []*partner{}
	}
	pathToPartnersfile := fmt.Sprintf("%s/files/%s", currentDir, partnersFileName)
	partnersFile, err := os.Open(pathToPartnersfile)
	if err != nil {
		log.Printf("failed to open partners file at [%s], erro %v", pathToPartnersfile, err)
		return []*partner{}
	}
	defer partnersFile.Close()
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		// Enforcing reading the TSE zip file as ISO 8859-1 (latin 1)
		r := csv.NewReader(charmap.ISO8859_1.NewDecoder().Reader(in))
		r.LazyQuotes = true
		r.Comma = ','
		return r
	})
	var partners []*partner
	if err := gocsv.UnmarshalFile(partnersFile, &partners); err != nil {
		log.Printf("failed to parse csv partner file to slice of partner struct, erro %v", err)
		return []*partner{}
	}
	log.Println("successfully loaded all partners from file")
	return partners
}
