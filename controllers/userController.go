package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fredele20/golang-jwt-project/database"
	"github.com/fredele20/golang-jwt-project/helpers"
	"github.com/fredele20/golang-jwt-project/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}

	return check, msg
}

func checkExistingUser(ctx context.Context, field, value string) (int64, error) {
	count, err := userCollection.CountDocuments(ctx, bson.M{field: value})
	if err != nil {
		log.Panic(err)
		fmt.Printf("Error checking for %v", field)
		return count, err
	}

	return count, nil
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		countEmail, _ := checkExistingUser(ctx, "email", *user.Email)
		countPhone, _ := checkExistingUser(ctx, "phone", *user.Phone)

		if countEmail > 0 || countPhone > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user with this phone number or email already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refereshToken, err := helpers.GenerateAuthToken(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *&user.User_id)
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while generating token"})
			return
		}
		user.Token = &token
		user.Refresh_token = &refereshToken

		_, insertError := userCollection.InsertOne(ctx, user)
		if insertError != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, user)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}

		validPassword, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if !validPassword {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}

		token, refreshToken, _ := helpers.GenerateAuthToken(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		helpers.UpdateAllToken(token, refreshToken, foundUser.User_id)
		err = userCollection.FindOne(c, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helpers.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page -1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}}, 
			{"total_count", bson.D{{"$sum", 1}}}, 
			{"data", bson.D{{"$push", "$$ROOT"}}},
		}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}},
		}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing users"})
		}

		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers[0])
		
	}
}

// only the admin has the acces to this request
// TODO: coming back here to implement this function better
// to allow users to get their informations but allows admin only to access other users info.
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		userEmail := c.GetString("email")
		userType := c.GetString("user_type")
		fmt.Println(userEmail)
		fmt.Println(userId)

		if userType != "ADMIN" {
			c.JSON(http.StatusBadGateway, gin.H{"error": "you can not do this"})
			return
		}

		if err := helpers.MatchUserTypeToUid(c, userId); err != nil {
			fmt.Println(userId)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			fmt.Println(err.Error())
			return
		}
		fmt.Println(userId)
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func GetUserById() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
