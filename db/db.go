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

// GetCandidateRoles returns a list with candidate roles
func (c *Client) GetCandidateRoles() ([]string, error) {
	var entities []*candidateType
	q := datastore.NewQuery(candidateRolesCollection)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to find all candidate roles from db on collection %s, error %v", candidateRolesCollection, err)
	}
	var roles []string
	for _, c := range entities {
		roles = append(roles, c.Role)
	}
	return roles, nil
}

// GetAvailableStates returns a list with availables states
func (c *Client) GetAvailableStates() ([]string, error) {
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

// GetCitiesOfState returns the city of a given state
func (c *Client) GetCitiesOfState(s string) ([]string, error) {
	var entities []*state
	q := datastore.NewQuery(statesCollection).Filter("State=", s)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, fmt.Errorf("failed to retrieve available cities of state %s from db on collection %s, erro %v", s, statesCollection, err)
	}
	return entities[0].Cities, nil
}
