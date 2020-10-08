package main

import (
	"context"
	"fmt"
	"time"

	"github.com/candidatos-info/descritor"
	pagination "github.com/gobeam/mongo-go-pagination"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

// func main() {
// 	// Establishing mongo db connection
// 	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
// 	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://api-user:243ksl2452msw4vna34k3da@cluster0.jt4t8.gcp.mongodb.net/candidatos"))
// 	if err != nil {
// 		log.Fatal("ROLA ", err)
// 	}
// 	limit := 5
// 	// index := 2
// 	findoptions := options.Find()
// 	if limit > 0 {
// 		findoptions.SetLimit(int64(limit))
// 		findoptions.SetSkip(int64(2))
// 		findoptions.SetSort(map[string]int{"transparency": -1})
// 	}
// 	cur, err := client.Database("candidatos").Collection("candidatures").Find(ctx, bson.M{}, findoptions)
// 	if err != nil {
// 		log.Fatal("PAU ", err)
// 	}
// 	defer cur.Close(context.Background())
// 	var data []descritor.CandidateForDB
// 	err = cur.All(context.Background(), &data)
// 	if err != nil {
// 		log.Fatal("BOSTA ", err)
// 	}
// 	for _, ss := range data {
// 		fmt.Println(ss.BallotName)
// 	}
// }

func main() {
	// Establishing mongo db connection
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://api-user:243ksl2452msw4vna34k3da@cluster0.jt4t8.gcp.mongodb.net/candidatos"))
	if err != nil {
		panic(err)
	}

	// Example for Normal Find query
	filter := bson.M{}
	var limit int64 = 3
	var page int64 = 5
	// Querying paginated data
	// Sort and select are optional
	paginatedData, err := pagination.New(client.Database("candidatos").Collection("candidatures")).Limit(limit).Page(page).Sort("transparency", -1).Filter(filter).Find()
	if err != nil {
		panic(err)
	}

	// paginated data is in paginatedData.Data
	// pagination info can be accessed in  paginatedData.Pagination
	// if you want to marshal data to your defined struct

	var lists []descritor.CandidateForDB
	for _, raw := range paginatedData.Data {
		var product *descritor.CandidateForDB
		if marshallErr := bson.Unmarshal(raw, &product); marshallErr == nil {
			lists = append(lists, *product)
		}

	}
	// print ProductList
	// fmt.Printf("Norm Find Data: %+v\n", lists)
	for _, s := range lists {
		fmt.Println(s.BallotName)
	}

	// print pagination data
	fmt.Printf("Normal find pagination info: %+v\n", paginatedData.Pagination)
}
