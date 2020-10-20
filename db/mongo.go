package db

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/exception"
	pagination "github.com/gobeam/mongo-go-pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	timeout = 15 // in seconds
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
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar estados disponíveis do banco na collection [%s], erro %q", descritor.LocationsCollection, err), nil)
	}
	sort.Strings(location.Cities) // TODO get it sorted from MongoDB?
	return location.Cities, nil
}

// GetCandidateByEmail searches for a candidate using email
func (c *Client) GetCandidateByEmail(email string, year int) (*descritor.CandidateForDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	var candidate descritor.CandidateForDB
	filter := bson.M{"email": strings.ToUpper(email), "year": year}
	if err := c.client.Database(c.dbName).Collection(descritor.CandidaturesCollection).FindOne(ctx, filter).Decode(&candidate); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar candidato pelo ano [%d] e pelo email [%s] no banco na collection [%s], erro %v", year, email, descritor.CandidaturesCollection, err), nil)
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
	filter := bson.M{
		"email": candidate.Email,
		"year":  candidate.Year,
	}
	update := bson.M{
		"$set": bson.M{
			"biography":      candidate.Biography,
			"transparency":   candidate.Transparency,
			"proposals":      candidate.Proposals,
			"contacts":       candidate.Contacts,
			"accepted_terms": candidate.AcceptedTerms,
		},
	}
	if _, err := c.client.Database(c.dbName).Collection(descritor.CandidaturesCollection).UpdateOne(ctx, filter, update); err != nil {
		return nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao atualizar perfil de candidato, erro %v", err), nil)
	}
	return candidate, nil
}

// FindTransparentCandidatures searches for a list of candidatures with proposals defined
func (c *Client) FindTransparentCandidatures(queryMap map[string]interface{}, pageSize, page int) ([]*descritor.CandidateForDB, *pagination.PaginationData, error) {
	query := make(bson.M, len(queryMap))
	for k, v := range queryMap {
		switch k {
		case "name":
			query["ballot_name"] = bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", queryMap["name"]), Options: "i"}}
		case "tags":
			if len(queryMap["tags"].([]string)) > 0 {
				query["proposals.topic"] = bson.M{"$in": queryMap["tags"]}
			}
		default:
			query[k] = v
		}
	}
	query["transparency"] = bson.M{"$gte": 0.0} // candidatures without proposals does not count!
	var candidatures []*descritor.CandidateForDB
	db := c.client.Database(c.dbName)
	p := pagination.New(db.Collection(descritor.CandidaturesCollection))
	paginatedData, err := p.Limit(int64(pageSize)).Page(int64(page)).Sort("transparency", -1).Filter(query).Find()
	if err != nil {
		return nil, nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar por lista candidatos, erro %v", err), nil)
	}
	for _, raw := range paginatedData.Data {
		var candidature *descritor.CandidateForDB
		if err := bson.Unmarshal(raw, &candidature); err != nil {
			return nil, nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao deserializar struct de candidatura a partir da resposta do banco, erro %v", err), nil)
		}
		candidatures = append(candidatures, candidature)
	}
	return candidatures, &paginatedData.Pagination, nil
}

// FindNonTransparentCandidatures searches for non transparent candidatures
func (c *Client) FindNonTransparentCandidatures(queryMap map[string]interface{}, pageSize, page int) ([]*descritor.CandidateForDB, *pagination.PaginationData, error) {
	query := make(bson.M, len(queryMap))
	for k, v := range queryMap {
		switch k {
		case "name":
			query["ballot_name"] = bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", queryMap["name"]), Options: "i"}}
		case "tags":
			continue
		default:
			query[k] = v
		}
	}
	query["transparency"] = bson.M{"$eq": nil}
	var candidatures []*descritor.CandidateForDB
	db := c.client.Database(c.dbName)
	p := pagination.New(db.Collection(descritor.CandidaturesCollection))
	paginatedData, err := p.Limit(int64(pageSize)).Page(int64(page)).Sort("transparency", -1).Filter(query).Find()
	if err != nil {
		return nil, nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar por lista candidatos, erro %v", err), nil)
	}
	for _, raw := range paginatedData.Data {
		var candidature *descritor.CandidateForDB
		if err := bson.Unmarshal(raw, &candidature); err != nil {
			return nil, nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao deserializar struct de candidatura a partir da resposta do banco, erro %v", err), nil)
		}
		candidatures = append(candidatures, candidature)
	}
	return candidatures, &paginatedData.Pagination, nil
}

// FindCandidatesWithParams searches for a list of candidates with given params
func (c *Client) FindCandidatesWithParams(queryMap map[string]interface{}, pageSize, page int) ([]*descritor.CandidateForDB, *pagination.PaginationData, error) {
	query := make(bson.M, len(queryMap))
	for k, v := range queryMap {
		switch k {
		case "name":
			query["ballot_name"] = bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", queryMap["name"]), Options: "i"}}
		case "tags":
			if len(queryMap["tags"].([]string)) > 0 {
				query["proposals.topic"] = bson.M{"$in": queryMap["tags"]}
			}
		default:
			query[k] = v
		}
	}
	var candidatures []*descritor.CandidateForDB
	db := c.client.Database(c.dbName)
	p := pagination.New(db.Collection(descritor.CandidaturesCollection))
	paginatedData, err := p.Limit(int64(pageSize)).Page(int64(page)).Sort("transparency", -1).Filter(query).Find()
	if err != nil {
		return nil, nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar por lista candidatos, erro %v", err), nil)
	}
	for _, raw := range paginatedData.Data {
		var candidature *descritor.CandidateForDB
		if err := bson.Unmarshal(raw, &candidature); err != nil {
			return nil, nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao deserializar struct de candidatura a partir da resposta do banco, erro %v", err), nil)
		}
		candidatures = append(candidatures, candidature)
	}
	return candidatures, &paginatedData.Pagination, nil
}

// FindRelatedCandidatesWithParams searches for a list of candidates with given params
func (c *Client) FindRelatedCandidatesWithParams(queryMap map[string]interface{}, pageSize, page int) ([]*descritor.CandidateForDB, *pagination.PaginationData, error) {
	query := make(bson.M, len(queryMap))
	for k, v := range queryMap {
		switch k {
		case "name":
			query["ballot_name"] = bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", queryMap["name"]), Options: "i"}}
		case "tags":
			if len(queryMap["tags"].([]string)) > 0 {
				query["proposals.topic"] = bson.M{"$in": queryMap["tags"]}
			}
		default:
			query[k] = v
		}
	}
	query["transparency"] = bson.M{"$gte": 0.0} // candidatures without proposals does not count!
	var candidatures []*descritor.CandidateForDB
	db := c.client.Database(c.dbName)
	p := pagination.New(db.Collection(descritor.CandidaturesCollection))
	paginatedData, err := p.Limit(int64(pageSize)).Page(int64(page)).Sort("transparency", -1).Filter(query).Find()
	if err != nil {
		return nil, nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao buscar por lista candidatos, erro %v", err), nil)
	}
	for _, raw := range paginatedData.Data {
		var candidature *descritor.CandidateForDB
		if err := bson.Unmarshal(raw, &candidature); err != nil {
			return nil, nil, exception.New(exception.NotFound, fmt.Sprintf("Falha ao deserializar struct de candidatura a partir da resposta do banco, erro %v", err), nil)
		}
		candidatures = append(candidatures, candidature)
	}
	return candidatures, &paginatedData.Pagination, nil
}
