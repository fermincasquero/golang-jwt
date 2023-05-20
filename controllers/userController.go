package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fredele20/golang-jwt-project/core"
	"github.com/fredele20/golang-jwt-project/database/mongod"
	"github.com/fredele20/golang-jwt-project/helpers"
	"github.com/fredele20/golang-jwt-project/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = mongod.UserCollection()
var validate = validator.New()

func checkExistingUser(ctx context.Context, field, value string) (int64, error) {
	count, err := userCollection.CountDocuments(ctx, bson.M{field: value})
	if err != nil {
		log.Panic(err)
		fmt.Printf("Error checking for %v", field)
		return count, err
	}

	return count, nil
}

func parsePhone(phone, iso2 string) (string, error) {
	num, err := phonenumbers.Parse(phone, iso2)
	if err != nil {
		return "", err
	}

	switch phonenumbers.GetNumberType(num) {
	case phonenumbers.VOIP, phonenumbers.VOICEMAIL:
		return "", errors.New("Sorry, this number can not be used")
	}

	return phonenumbers.Format(num, phonenumbers.E164), nil
}

// var con database.Store

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// if user.User_type == "" {
		// 	user.User_type = "USER"
		// }

		// validationErr := validate.Struct(user)
		// if validationErr != nil {
		// 	fmt.Println(user)
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		// 	return
		// }

		// phone, err := parsePhone(user.Phone, user.Iso2)
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "failed to validate phone"})
		// 	return
		// }

		// user.Phone = phone

		// countEmail, _ := checkExistingUser(ctx, "email", user.Email)
		// countPhone, _ := checkExistingUser(ctx, "phone", user.Phone)

		// if countEmail > 0 ||countPhone > 0 {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "user with this email or phone number already exists"})
		// 	return
		// }

		// password := HashPassword(user.Password)
		// user.Password = password

		// user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// user.ID = primitive.NewObjectID()
		// user.User_id = user.ID.Hex()

		// token, refereshToken, err := helpers.GenerateAuthToken(user.Email, user.FirstName, user.LastName, user.User_type, *&user.User_id)
		// if err != nil {
		// 	log.Panic(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while generating token"})
		// 	return
		// }
		// user.Token = &token
		// user.Refresh_token = &refereshToken

		// _, insertError := userCollection.InsertOne(ctx, user)
		// if insertError != nil {
		// 	msg := fmt.Sprintf("User item was not created")
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		// 	return
		// }

		// token, refereshToken, err := helpers.GenerateAuthToken(user.Email, user.FirstName, user.LastName, user.User_type, *&user.User_id)
		// if err != nil {
		// 	log.Panic(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while generating token"})
		// 	return
		// }
		// user.Token = &token
		// user.Refresh_token = &refereshToken

		newUser, err := core.CreateUser(ctx, user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, newUser)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		foundUser, err := core.Login(ctx, user.Email, user.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

func Logout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the token from the request header
		token := ctx.GetHeader("token")
		err := core.Logout(token)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, token)
	}
}

func ForgotPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx context.Context
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
			return
		}

		foundUser, err := core.ForgotPassword(ctx, user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

func ListUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx context.Context
		var filter models.ListUserFilter

		users, err := core.ListUsers(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helpers.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
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
