package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fredele20/golang-jwt-project/database"
	"github.com/fredele20/golang-jwt-project/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)


var productCollection *mongo.Collection = database.OpenCollection(database.Client, "product")


func AddProduct() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var productCtx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		var product models.Product

		if err := ctx.BindJSON(&product); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(product)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		// userId := ctx.GetString("uid")
		// userName := ctx.GetString("first_name")

		product.OwnerID = ctx.GetString("uid")
		product.OwnerName = ctx.GetString("first_name")
		product.ID = primitive.NewObjectID()
		product.Product_ID = product.ID.Hex()
		_, insertErr := productCollection.InsertOne(productCtx, product)
		if insertErr != nil {
			msg := fmt.Sprintf("Product item was not created")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		defer cancel()
		ctx.JSON(http.StatusCreated, product)
	}
}
