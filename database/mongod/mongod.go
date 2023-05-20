package mongod

import (
	"context"
	"errors"
	"os"

	"github.com/fredele20/golang-jwt-project/database"
	"github.com/fredele20/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dbStore struct {
	// client         *mongo.Client
	dbName         string
	collectionName string
}


var client *mongo.Client = database.DBInstance()

var dbName = os.Getenv("DATABASE_NAME")

func UserCollection() *mongo.Collection {
	return client.Database(dbName).Collection("user")
}

func SessionCollection() *mongo.Collection {
	return client.Database(dbName).Collection("session")
}

func ProdCollection() *mongo.Collection {
	return client.Database(dbName).Collection("product")
}

func PurchasedCollection() *mongo.Collection {
	return client.Database(dbName).Collection("purchasedProduct")
}

func GetUserByField(ctx context.Context, field, value string) (*models.User, error) {
	var user models.User
	if err := UserCollection().FindOne(ctx, bson.M{field: value}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	return GetUserByField(ctx, "phone", phone)
}

func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return GetUserByField(ctx, "email", email)
}

func GetUserById(ctx context.Context, id string) (*models.User, error) {
	return GetUserByField(ctx, "userId", id)
}

func ListUsers(ctx context.Context, filters models.ListUserFilter) (*models.UserList, error) {
	opts := options.Find()
	opts.SetProjection(bson.M{
		"password": false,
		"token": false,
	})
	
	if filters.Limit != 0 {
		opts.SetLimit(filters.Limit)
	}

	filter := bson.M{}

	if filters.Status != nil && filters.Status.IsValid() {
		filter["status"] = filters.Status.String()
	}

	var users []*models.User

	cursor, err := UserCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &users); err != nil {
		print(err)
		return nil, err
	}

	delete(filter, "id")

	count, err := UserCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &models.UserList{
		Count: count,
		Data: users,
	}, nil

}

func UpdateUser(ctx context.Context, payload *models.User) (*models.User, error) {
	var user models.User
	if err := UserCollection().FindOneAndUpdate(ctx, bson.M{"userId": payload.UserId}, bson.M{
		"$set": payload,
	}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func DeactivateUser(ctx context.Context, id string) (*models.User, error) {
	return UpdateUser(ctx, &models.User{UserId: id, Status: models.StatusDeactivated})
}

func ActivateUser(ctx context.Context, id string) (*models.User, error) {
	return UpdateUser(ctx, &models.User{UserId: id, Status: models.StatusActivated})
}

func ResetPassword(ctx context.Context, id, password string) (*models.User, error) {
	return UpdateUser(ctx, &models.User{UserId: id, Password: password})
}

func DeleteUser(ctx context.Context, id string) error {
	if _, err := UserCollection().DeleteOne(ctx, bson.M{"userId": id}); err != nil {
		return err
	}

	return nil
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
		return nil, ErrDuplicate
	}

	if _, err := UserCollection().InsertOne(ctx, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

var ErrDuplicate = errors.New("duplicate record")
