package db

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
	"github.com/candidatos-info/descritor"
)

var (
	rolesMap = map[string]string{
		"prefeito": "EM",
		"vereador": "LM",
	}
)

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
	var entities []*descritor.Location
	q := datastore.NewQuery(descritor.LocationsCollection)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to retrieve available states from db on collection %s, erro %v", descritor.LocationsCollection, err)
	}
	var states []string
	for _, s := range entities {
		states = append(states, s.State)
	}
	return states, nil
}

// GetCities returns the city of a given state
func (c *DataStoreClient) GetCities(s string) ([]string, error) {
	var entities []*descritor.Location
	q := datastore.NewQuery(descritor.LocationsCollection).Filter("state=", s)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to retrieve available cities of state %s from db on collection %s, erro %v", s, descritor.LocationsCollection, err)
	}
	return entities[0].Cities, nil
}

// FindCandidatesWithParams queries candidates with some given params
func (c *DataStoreClient) FindCandidatesWithParams(state, city, role string, year int) ([]*descritor.CandidateForDB, error) {
	var entities []*descritor.VotingCity
	q := datastore.NewQuery(descritor.CandidaturesCollection).Filter("year=", year).Filter("state=", state).Filter("city=", city).Filter("candidates.role=", rolesMap[role])
	if role == "prefeito" {
		q.Filter("candidates.role=", "VEM") // if prefeito is select this filter also selects for the vice
	}
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to find candidates for state [%s] and city [%s] and year [%d], erro %v", state, city, year, err)
	}
	if len(entities) == 0 {
		return []*descritor.CandidateForDB{}, nil
	}
	return entities[0].Candidates, nil
}

// GetCandidateBySequencialID searches for a candidate using its
// sequencial ID and returns it.
func (c *DataStoreClient) GetCandidateBySequencialID(year int, state, city, sequencialID string) (*descritor.CandidateForDB, error) {
	var entities []*descritor.VotingCity
	q := datastore.NewQuery(descritor.CandidaturesCollection).Filter("year=", year).Filter("state=", state).Filter("city=", city).Filter("candidates.sequencial_candidate=", sequencialID)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to find candidates for state [%s] and city [%s] and year [%d], erro %v", state, city, year, err)
	}
	return entities[0].Candidates[0], nil
}
