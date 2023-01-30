package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fredele20/golang-jwt-project/database"
	"github.com/fredele20/golang-jwt-project/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)


var productCollection *mongo.Collection = database.OpenCollection(database.Client, "product")
var purchasedProduct *mongo.Collection = database.OpenCollection(database.Client, "purchasedProduct")


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

		product.OwnerID = ctx.GetString("uid")
		product.OwnerName = ctx.GetString("first_name") + " " + ctx.GetString("last_name")
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

func GetProductById() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var productCtx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()
		productId := ctx.Param("product_id")

		var product models.Product

		err := productCollection.FindOne(productCtx, bson.M{"product_id": productId}).Decode(&product)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "product with the given ID is not found"})
			return
		}

		ctx.JSON(http.StatusOK, product)
	}
}


func GetProductsByOwnerId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var productCtx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()
		ownerId := ctx.Param("ownerid")

		var products []models.Product

		results, err := productCollection.Find(productCtx, bson.M{"ownerid": ownerId})
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "product with the owner id is not found"})
			return
		}

		defer results.Close(productCtx)
		for results.Next(productCtx) {
			var singleProduct models.Product
			if err = results.Decode(&singleProduct); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}

			products = append(products, singleProduct)
		}

		ctx.JSON(http.StatusOK, &products)
	}
}

func PurchaseProduct() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var productCtx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		var product models.Product
		var purchaseProduct models.PurchaseProduct
		if err := ctx.BindJSON(&purchaseProduct); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := productCollection.FindOne(productCtx, bson.M{"product_id": purchaseProduct.ProductId}).Decode(&product)
		fmt.Println(purchaseProduct.ProductId)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		_, insertErr := purchasedProduct.InsertOne(productCtx, purchaseProduct)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, purchaseProduct)
	}
}
