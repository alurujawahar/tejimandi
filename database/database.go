package database

import (
	"context"
	"fmt"
	"log"

	SmartApi "github.com/angel-one/smartapigo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
const mongoUrl = "mongodb://localhost:27017"

func QueryMongo(client *mongo.Client, tradingSymbol string) (SmartApi.OrderParams, bson.M){
	// Access a MongoDB collection
	collection := client.Database("stocks").Collection("list")

	// Define a filter for the query
	filter := bson.D{{Key: "tradingsymbol", Value: bson.D{{Key: "$eq", Value: tradingSymbol}}}}
	options := options.FindOne().SetProjection(bson.D{{Key: "_id", Value: 1}})

	// Define options for the query (e.g., sorting)
	// options := options.FindOne().SetSort(bson.D{{Key: "age", Value: 1}})
	// Find a single document in the collection based on the filter and options
	var objectId bson.M
	var result SmartApi.OrderParams
	err := collection.FindOne(context.Background(), filter, options).Decode(&objectId)
	if err == mongo.ErrNoDocuments {
		fmt.Println("No matching document found.")
	} else if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Query executed successfully!")
		fmt.Println(objectId["_id"])
	}

	err = collection.FindOne(context.Background(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Println("No matching document found.")
	} else if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Query executed successfully!")
	}
	return result, objectId
}

func UpdateMongo(client *mongo.Client, _id bson.M) {
	collection := client.Database("stocks").Collection("list")

    // Define the filter based on the document's _id
    filter := bson.D{{Key: "_id", Value: _id["_id"]}} // Replace with the actual _id

    // Define the update to be performed
    update := bson.D{
        {Key: "$set", Value: bson.D{{Key: "executed", Value: false}}}, 
    }

    // Perform the update
    result, err := collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Matched %v document(s) and modified %v document(s)\n", result.MatchedCount, result.ModifiedCount)
}

func ConnectMongo() *mongo.Client {
	clientOptions := options.Client().ApplyURI(mongoUrl)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return client 
}