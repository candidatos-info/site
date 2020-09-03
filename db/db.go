package db

import (
	"context"
	"fmt"

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

//Client manages all iteractions with mongodb
type Client struct {
	client *datastore.Client
}

// NewClient retuns a new Client
func NewClient(client *datastore.Client) *Client {
	return &Client{
		client: client,
	}
}

// GetStates returns a list with availables states
func (c *Client) GetStates() ([]string, error) {
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
func (c *Client) GetCities(s string) ([]string, error) {
	var entities []*state
	q := datastore.NewQuery(statesCollection).Filter("State=", s)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to retrieve available cities of state %s from db on collection %s, erro %v", s, statesCollection, err)
	}
	return entities[0].Cities, nil
}
