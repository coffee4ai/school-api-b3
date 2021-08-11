package mongodb

import (
	"context"
	"log"
	"time"

	"github.com/coffee4ai/school-api-b3/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBInstance struct {
	client             *mongo.Client
	userCollections    *mongo.Collection
	classCollections   *mongo.Collection
	sectionCollections *mongo.Collection
	roomCollections    *mongo.Collection
}

var schoolDB MongoDBInstance

func ConnectDB() (*mongo.Client, error) {

	url := config.GetMongoDbUrl()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}
	// defer client.Disconnect(ctx)
	schoolDB.client = client
	err = schoolDB.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(databases)
	db := schoolDB.client.Database(config.GetMongoDbName())

	schoolDB.userCollections = db.Collection("users")
	schoolDB.classCollections = db.Collection("class")
	schoolDB.sectionCollections = db.Collection("section")
	schoolDB.roomCollections = db.Collection("room")

	names, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	for _, n := range names {
		log.Println("Collection exist - ", n)
	}

	return schoolDB.client, nil
}

func GetCollectionHandle(name string) (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := schoolDB.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	var c *mongo.Collection
	switch name {
	case "users":
		c = schoolDB.userCollections
	case "class":
		c = schoolDB.classCollections
	case "section":
		c = schoolDB.sectionCollections
	case "room":
		c = schoolDB.roomCollections
	}
	return c, nil
}

func GetMongoIDFromHex(id interface{}) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id.(string))
}

func IsValidMongoID(id string) bool {
	return primitive.IsValidObjectID(id)
}

func GetStringFromMongoID(id interface{}) string {
	return id.(primitive.ObjectID).Hex()
}
