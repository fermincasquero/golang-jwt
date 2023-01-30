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
	ProductId string `json:"product_id" validate:"required"`
	Quantity  int    `json:"qty" validate:"required"`
}
