package mongod

import (
	"context"
	"errors"
	"os"

	"github.com/fredele20/golang-jwt-project/database"
	"github.com/fredele20/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type dbStore struct {
	// client         *mongo.Client
	dbName         string
	collectionName string
}

// func ConnectDBStore() (database.Store, error) {
// 	return &dbStore{}, nil
// }

// func DBConnections(connectionUri, databaseName string) (*mongo.Client, error){
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	opts := options.Client().ApplyURI(connectionUri)

// 	client, err := mongo.Connect(ctx, opts)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Println("Connected to Database")

// 	return client, nil
// }

// var Client *mongo.Client = DBConnections()

// func init() {
// 	ConnectDB()
// }

var client *mongo.Client = database.DBInstance()

var dbName = os.Getenv("DATABASE_NAME")

func UserCollection() *mongo.Collection {
	return client.Database(dbName).Collection("user")
}

func ProdCollection() *mongo.Collection {
	return client.Database(dbName).Collection("product")
}

func PurchasedCollection() *mongo.Collection {
	return client.Database(dbName).Collection("purchasedProduct")
}

func GetUserByField(ctx context.Context, field, value string) (models.User, error) {
	var user models.User
	if err := UserCollection().FindOne(ctx, bson.M{field: value}).Decode(&user); err != nil {
		return user, err
	}
	return user, nil
}

func GetUserByPhone(ctx context.Context, phone string) (models.User, error) {
	return GetUserByField(ctx, "phone", phone)
}

func GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	return GetUserByField(ctx, "email", email)
}

func CreateUser(ctx context.Context, payload *models.User) (*models.User, error) {
	filters := bson.M{
		"$or": []bson.M{
			{
				"email": payload.Email,
			},
			{
				"phone": payload.Phone,
			},
		},
	}

	var user models.User

	if err := UserCollection().FindOne(ctx, filters).Decode(&user); err == nil {
		return nil, err
	}

	if _, err := UserCollection().InsertOne(ctx, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

var ErrDuplicate = errors.New("duplicate record")

// func (d dbStore) UpdateOne(ctx context.Context, filtre, object, opts interface{}) (*models.User, error) {

// }
