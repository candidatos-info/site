package db

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
)

var (
	candidateRolesCollection = "candidateRoles"
	statesCollection         = "states"
)

// CandidateForDB is a struct with a fields portion of descritor.Candidature. This is struct
// is used only for DB purposes.
type CandidateForDB struct {
	SequencialCandidate string `datastore:"sequencial_candidate,omitempty"` // Sequencial code of candidate on TSE system.
	Site                string `datastore:"site,omitempty"`                 // Site of candidate.
	Facebook            string `datastore:"facebook,omitempty"`             // Facebook of candidate.
	Twitter             string `datastore:"twitter,omitempty"`              // Twitter of candidate.
	Instagram           string `datastore:"instagram,omitempty"`            // Instagram of candidate.
	Description         string `datastore:"description,omitempty"`          // Description of candidate.
	Biography           string `datastore:"biography,omitempty"`            // Biography of candidate.
	PhotoURL            string `datastore:"photo_url,omitempty"`            // Photo URL of candidate.
	LegalCode           string `datastore:"legal_code,omitempty"`           // Brazilian Legal Code (CPF) of candidate.
	Party               string `datastore:"party,omitempty"`                // Party of candidate.
	Name                string `datastore:"name,omitempty"`                 // Natural name of candidate.
	BallotName          string `datastore:"ballot_name,omitempty"`          // Ballot name of candidate.
	BallotNumber        int    `datastore:"ballot_number,omitempty"`        // Ballot number of candidate.
	Email               string `datastore:"email,omitempty"`                // Email of candidate.
}

// db schema
type votingCity struct {
	City       string
	State      string
	Candidates []*CandidateForDB
}

type candidateType struct {
	Role string
}

type state struct {
	State  string
	Cities []string
}

//DataStoreClient manages all iteractions with mongodb
type DataStoreClient struct {
	client *datastore.Client
}

// NewDataStoreClient retuns a new Client
func NewDataStoreClient(gcpProjectID string) *DataStoreClient {
	client, err := datastore.NewClient(context.Background(), gcpProjectID)
	if err != nil {
		log.Fatalf("falha ao criar cliente do Datastore, erro %q", err)
	}
	return &DataStoreClient{
		client: client,
	}
}

// GetStates returns a list with availables states
func (c *DataStoreClient) GetStates() ([]string, error) {
	var entities []*state
	q := datastore.NewQuery(statesCollection)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to retrieve available states from db on collection %s, erro %v", statesCollection, err)
	}
	var states []string
	for _, s := range entities {
		states = append(states, s.State)
	}
	return states, nil
}

// GetCities returns the city of a given state
func (c *DataStoreClient) GetCities(s string) ([]string, error) {
	// var entities []*state
	// q := datastore.NewQuery(statesCollection).Filter("State=", s)
	// if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
	// 	return nil, fmt.Errorf("failed to retrieve available cities of state %s from db on collection %s, erro %v", s, statesCollection, err)
	// }
	// return entities[0].Cities, nil
	return []string{"Maceio", "Capela", "Penado", "Atalaia"}, nil
}

// FindCandidatesWithParams queries candidates with some given params
func (c *DataStoreClient) FindCandidatesWithParams(state, city, role string, year int) ([]*CandidateForDB, error) {
	return []*CandidateForDB{
		{
			SequencialCandidate: "505",
			Name:                "Aurelio",
			BallotName:          "Lelinho da Saem",
			BallotNumber:        505,
			Twitter:             "abmf",
			Party:               "PT",
			PhotoURL:            "https://upload.wikimedia.org/wikipedia/commons/7/7f/Daniel_Craig_-_Film_Premiere_%22Spectre%22_007_-_on_the_Red_Carpet_in_Berlin_%2822387409720%29_%28cropped%29.jpg",
		},
		{
			SequencialCandidate: "501",
			Name:                "Aurelio",
			BallotName:          "Lelinho da Saem",
			BallotNumber:        505,
			Twitter:             "abmf",
			Party:               "PT",
			PhotoURL:            "https://upload.wikimedia.org/wikipedia/commons/7/7f/Daniel_Craig_-_Film_Premiere_%22Spectre%22_007_-_on_the_Red_Carpet_in_Berlin_%2822387409720%29_%28cropped%29.jpg",
		},
		{
			SequencialCandidate: "502",
			Name:                "Aurelio",
			BallotName:          "Lelinho da Saem",
			BallotNumber:        505,
			Twitter:             "abmf",
			Party:               "PT",
			PhotoURL:            "https://upload.wikimedia.org/wikipedia/commons/7/7f/Daniel_Craig_-_Film_Premiere_%22Spectre%22_007_-_on_the_Red_Carpet_in_Berlin_%2822387409720%29_%28cropped%29.jpg",
		},
		{
			SequencialCandidate: "503",
			Name:                "Aurelio",
			BallotName:          "Lelinho da Saem",
			BallotNumber:        505,
			Twitter:             "abmf",
			Party:               "PT",
			PhotoURL:            "https://upload.wikimedia.org/wikipedia/commons/7/7f/Daniel_Craig_-_Film_Premiere_%22Spectre%22_007_-_on_the_Red_Carpet_in_Berlin_%2822387409720%29_%28cropped%29.jpg",
		},
		{
			SequencialCandidate: "505",
			Name:                "Aurelio",
			BallotName:          "Lelinho da Saem",
			BallotNumber:        505,
			Twitter:             "abmf",
			Party:               "PT",
			PhotoURL:            "https://upload.wikimedia.org/wikipedia/commons/7/7f/Daniel_Craig_-_Film_Premiere_%22Spectre%22_007_-_on_the_Red_Carpet_in_Berlin_%2822387409720%29_%28cropped%29.jpg",
		},
		{
			SequencialCandidate: "5090",
			Name:                "Aurelio",
			BallotName:          "Lelinho da Saem",
			BallotNumber:        505,
			Twitter:             "abmf",
			Party:               "PT",
			PhotoURL:            "https://upload.wikimedia.org/wikipedia/commons/7/7f/Daniel_Craig_-_Film_Premiere_%22Spectre%22_007_-_on_the_Red_Carpet_in_Berlin_%2822387409720%29_%28cropped%29.jpg",
		},
	}, nil
}

// GetCandidateBySequencialID searches for a candidate using its
// sequencial ID and returns it.
func (c *DataStoreClient) GetCandidateBySequencialID(year int, state, city, sequencialID string) (*CandidateForDB, error) {
	return &CandidateForDB{SequencialCandidate: "505",
		Name:         "Aurelio",
		BallotName:   "Lelinho da Saem",
		BallotNumber: 505,
		Twitter:      "abmf",
		Party:        "PT",
		PhotoURL:     "https://upload.wikimedia.org/wikipedia/commons/7/7f/Daniel_Craig_-_Film_Premiere_%22Spectre%22_007_-_on_the_Red_Carpet_in_Berlin_%2822387409720%29_%28cropped%29.jpg",
		Description:  "viado que da o cu e chupa pau",
	}, nil
}
