package db

import (
	"fmt"
	"strings"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/exception"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Client manages all iteractions with mongodb
type Client struct {
	client *mgo.Database
	dbName string
}

//NewMongoClient returns an db connection instance that can be used for CRUD opetations
func NewMongoClient(url, database string) (*Client, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: session.DB(database),
		dbName: database,
	}, nil
}

// GetStates returns a list of available states
func (c *Client) GetStates() ([]string, error) {
	var locations []descritor.Location
	if err := c.client.C(descritor.LocationsCollection).Find(bson.M{}).All(&locations); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco na collection [%s], erro %v", descritor.LocationsCollection, err), nil)
	}
	var states []string
	for _, location := range locations {
		states = append(states, location.State)
	}
	return states, nil
}

// GetCities returns the city of a given state
func (c *Client) GetCities(state string) ([]string, error) {
	var location descritor.Location
	if err := c.client.C(descritor.LocationsCollection).Find(bson.M{"state": state}).One(&location); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco na collection [%s], erro %v", descritor.LocationsCollection, err), nil)
	}
	return location.Cities, nil
}

// GetCandidateByEmail searches for a candidate using email
func (c *Client) GetCandidateByEmail(email string, year int) (*descritor.CandidateForDB, error) {
	var candidate descritor.CandidateForDB
	if err := c.client.C(descritor.CandidaturesCollection).Find(bson.M{"email": email, "year": year}).One(&candidate); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco na collection [%s], erro %v", descritor.LocationsCollection, err), nil)
	}
	return &candidate, nil
}

// FindCandidateBySequencialIDAndYear searches for a candidate using its
// sequencial ID and returns it.
func (c *Client) FindCandidateBySequencialIDAndYear(year int, sequencialID string) (*descritor.CandidateForDB, error) {
	var candidate descritor.CandidateForDB
	if err := c.client.C(descritor.CandidaturesCollection).Find(bson.M{"sequencial_candidate": sequencialID, "year": year}).One(&candidate); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco na collection [%s], erro %v", descritor.LocationsCollection, err), nil)
	}
	return &candidate, nil
}

// UpdateCandidateProfile updates the profile of a cndidate
func (c *Client) UpdateCandidateProfile(candidate *descritor.CandidateForDB) (*descritor.CandidateForDB, error) {
	if err := c.client.C(descritor.CandidaturesCollection).Update(bson.M{"email": candidate.Email, "year": candidate.Year}, bson.M{"$set": bson.M{"biography": candidate.Biography, "transparency": candidate.Transparence, "description": candidate.Description, "tags": candidate.Tags, "contact": candidate.Contact}}); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao atualizar perfil de candidato, erro %v", err), nil)
	}
	return candidate, nil
}

// FindCandidatesWithParams searches for a list of candidates with given params
func (c *Client) FindCandidatesWithParams(year int, state, city, role, gender string, tags []string, name string) ([]*descritor.CandidateForDB, error) {
	var candidates []*descritor.CandidateForDB
	queryMap := make(map[string]interface{})
	queryMap["year"] = year
	queryMap["state"] = state
	if city != "" {
		queryMap["city"] = city
	}
	if role != "" {
		queryMap["role"] = role
	}
	if gender != "" {
		queryMap["gender"] = gender
	}
	if name != "" {
		queryMap["name"] = name
	}
	if len(tags) != 0 {
		queryMap["tags"] = tags
	}
	sortBy := []string{"-transparency"}
	if err := c.client.C(descritor.CandidaturesCollection).Find(resolveQuery(queryMap)).Sort(strings.Join(sortBy, ",")).All(&candidates); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco na collection [%s], erro %v", descritor.LocationsCollection, err), nil)
	}
	return candidates, nil
}

func resolveQuery(query map[string]interface{}) bson.M {
	result := make(bson.M, len(query))
	for k, v := range query {
		switch k {
		case "name":
			result["ballot_name"] = bson.M{"$regex": bson.RegEx{Pattern: fmt.Sprintf(".*%s.*", query["name"]), Options: "i"}}
		case "tags":
			result["tags"] = bson.M{"$in": query["tags"]}
		default:
			result[k] = v
		}
	}
	return result
}
