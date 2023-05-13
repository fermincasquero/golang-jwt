package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fredele20/golang-jwt-project/database/mongod"
	"github.com/fredele20/golang-jwt-project/models"
	"github.com/nyaruka/phonenumbers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrCreateUserFailed              = errors.New("failed to create user")
	ErrCreateUserDuplicate           = errors.New("failed to create user because a user with this credentials already exists")
	ErrUserValidationFailed          = errors.New("failed to validate user before persisting")
	ErrListUsersFailed               = errors.New("failed to list users")
	ErrUpdateUserFailed              = errors.New("failed to update user")
	ErrDeleteUserFailed              = errors.New("failed to delete user")
	ErrUserDeactivationFailed        = errors.New("failed to deactivate user")
	ErrUserActivationFailed          = errors.New("failed to activate user")
	ErrAuthenticationFailed          = errors.New("Sorry, email/password incorrect. Please try again.")
	ErrAuthFailedAccountDeactivated  = errors.New("failed to authenticate user, account has been deactivated")
	ErrUserNotFoundById              = errors.New("user not found by id")
	ErrUserNotFoundByEmail           = errors.New("user not found by email")
	ErrUserNotFoundByPhone           = errors.New("user not found by phone")
	ErrFailedtoGetUserByEmail        = errors.New("Sorry, incorrect email. Please try again.")
	ErrFailedToResetPassword         = errors.New("Failed to rest password")
	ErrFailedToResetPasswordBadToken = errors.New("Sorry, your reset token has expired. Please try requesting for password reset again.")
	ErrPasswordIsSame                = errors.New("You cannot use this password, please login")
	ErrPasswordDoesNotMatch          = errors.New("Password does not match, please try again")

	ErrEmailDoesNotExist = errors.New("Email address does not exist")
)

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

func CreateUser(ctx context.Context, payload models.User) (*models.User, error) {
	if err := payload.Validate(); err != nil {
		return nil, err
	}

	phone, err := parsePhone(payload.Phone, payload.Iso2)
	if err != nil {
		return nil, err
	}

	payload.Phone = phone

	payload.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	payload.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	payload.ID = primitive.NewObjectID()
	payload.User_id = payload.ID.Hex()

	user, err := mongod.CreateUser(ctx, &payload)
	if err != nil {
		fmt.Println(err.Error())
		if err == mongod.ErrDuplicate {
			return nil, ErrCreateUserDuplicate
		}
		return nil, ErrCreateUserFailed
	}

	return user, nil
}
