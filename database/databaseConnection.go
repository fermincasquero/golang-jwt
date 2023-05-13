package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store interface {
	// GetUserByPhone(ctx context.Context, phone string) (models.User, error)
	// UpdateOne(ctx context.Context, filtre, object, opts interface{}) (*models.User, error)
}

func DBInstance() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	MongoDb := os.Getenv("DATABASE_URL")

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDb))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connected to MongoDB...")

	return client
}

// var Client *mongo.Client = DBInstance()

// func userColl() *mongo.Collection {
// 	return Client.Database("").Collection("")
// }

// func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
// 	dbName := os.Getenv("DATABASE_NAME")

// 	var collection *mongo.Collection = client.Database(dbName).Collection(collectionName)
// 	return collection
// }
