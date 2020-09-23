package db

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/exception"
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
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco %s, erro %v", descritor.LocationsCollection, err), nil)
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
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar cidades do estado %s do banco da coleção %s, erro %v", s, descritor.LocationsCollection, err), nil)
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
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidato por estado [%s] e cidade [%s] e ano [%d], erro %v", state, city, year, err), nil)
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
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidato por estado [%s] e cidade [%s] e ano [%d], erro %v", state, city, year, err), nil)
	}
	if len(entities) == 0 {
		return nil, exception.New(exception.NotFound, "Falha ao buscar candidato usando código sequencial", nil)
	}
	if len(entities[0].Candidates) == 0 {
		return nil, exception.New(exception.NotFound, "Falha ao buscar candidato usando código sequencial", nil)
	}
	return entities[0].Candidates[0], nil
}

// FindCandidateBySequencialIDAndYear searches for a candidate using its
// sequencial ID and returns it.
func (c *DataStoreClient) FindCandidateBySequencialIDAndYear(year int, sequencialID string) (*descritor.CandidateForDB, error) {
	var entities []*descritor.VotingCity
	q := datastore.NewQuery(descritor.CandidaturesCollection).Filter("year=", year).Filter("candidates.sequencial_candidate=", sequencialID)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidato por ano [%d] e sequencial id [%s], erro %v", year, sequencialID, err), nil)
	}
	if len(entities) == 0 {
		return nil, exception.New(exception.NotFound, "Falha ao buscar candidato usando código sequencial", nil)
	}
	if len(entities[0].Candidates) == 0 {
		return nil, exception.New(exception.NotFound, "Falha ao buscar candidato usando código sequencial", nil)
	}
	return entities[0].Candidates[0], nil
}

// GetCandidateByEmail searches for a candidate using email
func (c *DataStoreClient) GetCandidateByEmail(email string, year int) (*descritor.CandidateForDB, error) {
	var entities []*descritor.VotingCity
	q := datastore.NewQuery(descritor.CandidaturesCollection).Filter("candidates.email=", email).Filter("year=", year)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidato por email [%s], erro %v", email, err), nil)
	}
	if len(entities) == 0 {
		return nil, exception.New(exception.NotFound, "Email não cadastrado", nil)
	}
	if len(entities[0].Candidates) == 0 {
		return nil, exception.New(exception.NotFound, "Email não cadastrado", nil)
	}
	return entities[0].Candidates[0], nil
}

// GetVotingCityByCandidateEmail searches for a voting city using a candidate email
func (c *DataStoreClient) GetVotingCityByCandidateEmail(email string) (*descritor.VotingCity, error) {
	var entities []*descritor.VotingCity
	q := datastore.NewQuery(descritor.CandidaturesCollection).Filter("candidates.email=", email)
	if _, err := c.client.GetAll(context.Background(), q, &entities); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar local de votação por email de candidato [%s], erro %v", email, err), nil)
	}
	if len(entities) == 0 {
		return nil, exception.New(exception.NotFound, "Email não cadastrado", nil)
	}
	if len(entities[0].Candidates) == 0 {
		return nil, exception.New(exception.NotFound, "Email não cadastrado", nil)
	}
	return entities[0], nil
}

// UpdateVotingCity updates a voting city
func (c *DataStoreClient) UpdateVotingCity(votingCity *descritor.VotingCity) (*descritor.VotingCity, error) {
	key := datastore.NameKey(descritor.CandidaturesCollection, fmt.Sprintf("%s_%s", votingCity.State, votingCity.City), nil)
	if _, err := c.client.Put(context.Background(), key, votingCity); err != nil {
		return nil, exception.New(exception.ProcessmentError, fmt.Sprintf("Falha ao atualizar voting city, erro %v", err), nil)
	}
	return votingCity, nil
}
