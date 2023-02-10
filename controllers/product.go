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
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		var product models.Product

		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(product)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		product.OwnerID = c.GetString("uid")
		product.OwnerName = c.GetString("first_name") + " " + c.GetString("last_name")
		product.ID = primitive.NewObjectID()
		product.Product_ID = product.ID.Hex()
		_, insertErr := productCollection.InsertOne(ctx, product)
		if insertErr != nil {
			msg := fmt.Sprintf("Product item was not created")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusCreated, product)
	}
}

func GetProductById() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()
		productId := c.Param("product_id")

		var product models.Product

		err := productCollection.FindOne(ctx, bson.M{"product_id": productId}).Decode(&product)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "product with the given ID is not found"})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}


func GetProductsByOwnerId() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()
		ownerId := c.Param("ownerid")

		var products []models.Product

		results, err := productCollection.Find(ctx, bson.M{"ownerid": ownerId})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "product with the owner id is not found"})
			return
		}

		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleProduct models.Product
			if err = results.Decode(&singleProduct); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}

			products = append(products, singleProduct)
		}

		c.JSON(http.StatusOK, &products)
	}
}


// TODO: Going to call PDF generation functionality to this
// and also sending of emails upon every successful transaction
func PurchaseProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
		defer cancel()

		var product models.Product
		var purchaseProduct models.PurchaseProduct
		if err := c.BindJSON(&purchaseProduct); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := productCollection.FindOne(ctx, bson.M{"product_id": purchaseProduct.ProductId}).Decode(&product)
		fmt.Println(purchaseProduct.ProductId)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		if product.Quantity < purchaseProduct.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough item available, go for a lesser one."})
			return
		}

		purchaseProduct.SellerId = product.OwnerID
		purchaseProduct.ProductName = *product.Name
		purchaseProduct.SellerName = product.OwnerName
		purchaseProduct.BuyerId = c.GetString("uid")
		purchaseProduct.BuyerName = c.GetString("first_name") + " " + c.GetString("last_name")
		purchaseProduct.TransactionDate, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		_, insertErr := purchasedProduct.InsertOne(ctx, purchaseProduct)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		updateProductQty := bson.M{"quantity": product.Quantity - purchaseProduct.Quantity}
		_, err = productCollection.UpdateOne(ctx, bson.M{"product_id": product.Product_ID}, bson.M{"$set": updateProductQty})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, purchaseProduct)
	}
}
