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
	var entities []*state
	q := datastore.NewQuery(statesCollection).Filter("State=", s)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to retrieve available cities of state %s from db on collection %s, erro %v", s, statesCollection, err)
	}
	return entities[0].Cities, nil
}
