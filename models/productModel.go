package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID `bson:"id"`
	Name        *string            `json:"name" validate:"required,min=3,max=255"`
	Description *string            `json:"description" validate:"required,min=3,max=255"`
	Price       *string            `json:"price" validate:"required"`
	Quantity    int                `json:"qty" validate:"required,min=1"`
	OwnerID     string             `json:"owner_id"`
	OwnerName   string             `json:"owner_name"`
	Created_at  time.Time          `json:"created_at"`
	Updated_at  time.Time          `json:"updated_at"`
	Product_ID  string             `json:"product_id"`
}

type PurchaseProduct struct {
	ProductId       string    `json:"product_id" validate:"required"`
	ProductName     string    `json:"product_name"`
	Quantity        int       `json:"qty" validate:"required"`
	SellerId        string    `json:"seller_id"`
	SellerName      string    `json:"seller_name"`
	BuyerId         string    `json:"buyer_id"`
	BuyerName       string    `json:"buyer_name"`
	TransactionDate time.Time `json:"transaction_date"`
}
