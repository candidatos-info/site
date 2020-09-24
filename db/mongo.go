package db

import (
	"fmt"

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
