package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                 primitive.ObjectID `bson:"id"`
	FirstName          string             `json:"firstName" validate:"required,min=2,max=100"`
	LastName           string             `json:"lastName" validate:"required,min=2,max=100"`
	Password           string             `json:"password" validate:"required,min=6"`
	Email              string             `json:"email"`
	Phone              string             `json:"phone"`
	Token              *string            `json:"token"`
	Iso2               string             `json:"iso2" validate:"required"`
	Country            string             `json:"country"`
	UserType           string             `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	RefreshToken       *string            `json:"refreshToken"`
	ResetPasswordToken *string            `json:"resetPasswordToken"`
	CreatedAt          time.Time          `json:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt"`
	UserId             string             `json:"userId"`
	Status             Status             `json:"status"`
}

type Status string

const (
	StatusActivated   Status = "activated"
	StatusDeactivated Status = "deactivated"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActivated, StatusDeactivated:
		return true
	default:
		return false
	}
}

func (s Status) String() string {
	return string(s)
}

func (u User) Validate() error {
	if err := validation.ValidateStruct(&u,
		validation.Field(&u.FirstName, validation.Required),
		validation.Field(&u.LastName, validation.Required),
		// validation.Field(&u.Email, validation.Required, is.Email),
		// validation.Field(&u.Phone, validation.Required, is.E164),
		validation.Field(&u.Iso2, validation.Required, is.CountryCode2),
		validation.Field(&u.Country, validation.Required),
	); err != nil {
		return err
	}

	return nil
}

func (u User) ValidatePhone() error {
	return validation.Validate(u.Phone, is.E164)
}

type UserList struct {
	Data  []*User `json:"data"`
	Count int64   `json:"count"`
}

type ListUserFilter struct {
	// NextCursorId is used to paginate forward
	NextCursorId *string `json:"nextCursorId"`
	// PreviousCursorId is used to paginate backward
	PreviousCursorId *string `json:"previousCursorId"`
	// Filters by status
	Status *Status `json:""`
	// Filter by country
	Iso2 *string `json:"iso2"`
	// Limit the number of records to be returned at once
	Limit int64 `json:"limit"`
}

type ConfirmPasswordRequest struct {
	Password string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}
