package db

import (
	"context"
	"fmt"
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/exception"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

const (
	timeout = 10 // in seconds
)

//Client manages all iteractions with mongodb
type Client struct {
	client *mongo.Client
	dbName string
}

//NewMongoClient returns an db connection instance that can be used for CRUD opetations
func NewMongoClient(dbURL, dbName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB at link [%s], error %v", dbURL, err)
	}
	return &Client{
		client: client,
		dbName: dbName,
	}, nil
}

// GetStates returns a list of available states
func (c *Client) GetStates() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	var locations []*descritor.Location
	filter := bson.M{}
	cursor, err := c.client.Database(c.dbName).Collection(descritor.LocationsCollection).Find(ctx, filter, nil)
	if err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco na collection [%s], erro %v", descritor.LocationsCollection, err), nil)
	}
	if err = cursor.All(ctx, &locations); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	var location descritor.Location
	filter := bson.M{"state": state}
	if err := c.client.Database(c.dbName).Collection(descritor.LocationsCollection).FindOne(ctx, filter).Decode(&location); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco na collection [%s], erro %v", descritor.LocationsCollection, err), nil)
	}
	return location.Cities, nil
}

// GetCandidateByEmail searches for a candidate using email
func (c *Client) GetCandidateByEmail(email string, year int) (*descritor.CandidateForDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	var candidate descritor.CandidateForDB
	filter := bson.M{"email": email, "year": year}
	if err := c.client.Database(c.dbName).Collection(descritor.CandidaturesCollection).FindOne(ctx, filter).Decode(&candidate); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidato pelo ano [%d] e pelo email [%s] no banco na collection [%s], erro %v", year, email, descritor.LocationsCollection, err), nil)
	}
	return &candidate, nil
}

// FindCandidateBySequencialIDAndYear searches for a candidate using its
// sequencial ID and returns it.
func (c *Client) FindCandidateBySequencialIDAndYear(year int, sequencialID string) (*descritor.CandidateForDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	var candidate descritor.CandidateForDB
	filter := bson.M{"sequencial_candidate": sequencialID, "year": year}
	if err := c.client.Database(c.dbName).Collection(descritor.CandidaturesCollection).FindOne(ctx, filter).Decode(&candidate); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidato pelo ano [%d] e pelo sequencial ID [%s] no banco na collection [%s], erro %v", year, sequencialID, descritor.LocationsCollection, err), nil)
	}
	return &candidate, nil
}

// UpdateCandidateProfile updates the profile of a cndidate
func (c *Client) UpdateCandidateProfile(candidate *descritor.CandidateForDB) (*descritor.CandidateForDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	filter := bson.M{"email": candidate.Email, "year": candidate.Year}
	update := bson.M{"$set": bson.M{"biography": candidate.Biography, "transparency": candidate.Transparency, "description": candidate.Description, "tags": candidate.Tags, "contact": candidate.Contact}}
	if _, err := c.client.Database(c.dbName).Collection(descritor.CandidaturesCollection).UpdateOne(ctx, filter, update); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao atualizar perfil de candidato, erro %v", err), nil)
	}
	return candidate, nil
}

// FindCandidatesWithParams searches for a list of candidates with given params
func (c *Client) FindCandidatesWithParams(year int, state, city, role, gender string, tags []string, name string) ([]*descritor.CandidateForDB, error) {
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
	findOptions := options.Find()
	findOptions.SetSort(map[string]int{"transparency": -1})
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	var candidatures []*descritor.CandidateForDB
	cursor, err := c.client.Database(c.dbName).Collection(descritor.CandidaturesCollection).Find(ctx, resolveQuery(queryMap), findOptions)
	if err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidatos na collection [%s], erro %v", descritor.CandidaturesCollection, err), nil)
	}
	if err = cursor.All(ctx, &candidatures); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidatos na collection [%s], erro %v", descritor.CandidaturesCollection, err), nil)
	}
	return candidatures, nil
}

func resolveQuery(query map[string]interface{}) bson.M {
	result := make(bson.M, len(query))
	for k, v := range query {
		switch k {
		case "name":
			result["ballot_name"] = bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", query["name"]), Options: "i"}}
		case "tags":
			result["tags"] = bson.M{"$in": query["tags"]}
		default:
			result[k] = v
		}
	}
	return result
}
