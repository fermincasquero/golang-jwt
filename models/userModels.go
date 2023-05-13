package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"id"`
	FirstName     string            `json:"firstName" validate:"required,min=2,max=100"`
	LastName      string            `json:"lastName" validate:"required,min=2,max=100"`
	Password      string            `json:"password" validate:"required,min=6"`
	Email         string            `json:"email" validate:"email,required"`
	Phone         string             `json:"phone"`
	Token         *string            `json:"token"`
	Iso2          string             `json:"iso2" validate:"required"`
	User_type     string             `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
}

func (u User) Validate() error {
	if err := validation.ValidateStruct(&u,
		validation.Field(&u.FirstName, validation.Required),
		validation.Field(&u.LastName, validation.Required),
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Phone, validation.Required, is.E164),
		validation.Field(&u.Iso2, validation.Required, is.CountryCode2),
		// validation.Field(&u.Country, validation.Required),
	); err != nil {
		return err
	}

	return nil
}

func (u User) ValidatePhone() error {
	return validation.Validate(u.Phone, is.E164)
}
