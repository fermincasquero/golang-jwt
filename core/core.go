package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fredele20/golang-jwt-project/database/mongod"
	"github.com/fredele20/golang-jwt-project/libs/session"
	"github.com/fredele20/golang-jwt-project/models"
	"github.com/fredele20/golang-jwt-project/utils"
	"github.com/nyaruka/phonenumbers"
	"github.com/sirupsen/logrus"
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
	ErrFailedToGetUserByEmail        = errors.New("Sorry, incorrect email. Please try again.")
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

func buildPictureURLFromName(name string) string {
	return fmt.Sprintf("https://ui-avatars.com/api/?name=%s", strings.ReplaceAll(name, " ", "+"))
}

func CreateUser(ctx context.Context, payload models.User) (*models.User, error) {
	if err := payload.Validate(); err != nil {
		logrus.WithError(err).Error(ErrUserValidationFailed.Error())
		return nil, err
	}

	phone, err := parsePhone(payload.Phone, payload.Iso2)
	if err != nil {
		logrus.WithError(err).Error("failed to validate phone number or country code")
		return nil, err
	}

	
	payload.Phone = phone
	password := utils.HashPassword(payload.Password)
	payload.Password = password
	
	if payload.UserType == "" {
		payload.UserType = "USER"
	}
	
	payload.Status = models.StatusActivated
	payload.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	payload.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	payload.ID = primitive.NewObjectID()
	payload.UserId = payload.ID.Hex()

	
	user, err := mongod.CreateUser(ctx, &payload)
	if err != nil {
		fmt.Println(err.Error())
		if err == mongod.ErrDuplicate {
			logrus.WithError(err).Error("create user failed, duplicate record attempted")
			return nil, ErrCreateUserDuplicate
		}
		logrus.WithError(err).Error(err.Error())
		return nil, ErrCreateUserFailed
	}

	// token, err := session.CreateSession(session.Session{
	// 	AccountId: user.UserId,
	// 	Role: user.UserType,
	// 	Validity: 1,
	// 	UnitOfValidity: session.UnitOfValidityHour,
	// })

	if err != nil {
		logrus.WithError(err).Error("failed go generate authentication token")
		return nil, err
	}

	// user.Token = &token


	return user, nil
}

func Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := mongod.GetUserByEmail(ctx, email)
	if err != nil {
		logrus.WithError(err).Error("failed to get user by email")
		return nil, ErrAuthenticationFailed
	}

	validPassword, _ := utils.VerifyPassword(user.Password, password)
	if !validPassword {
		logrus.WithError(err).Error("failed to log user in, incorrect password")
		return nil, ErrAuthenticationFailed
	}

	token, err := session.CreateSession(session.Session{
		AccountId: user.UserId,
		Role: user.UserType,
		Validity: 1,
		UnitOfValidity: session.UnitOfValidityHour,
	})

	user.Token = &token

	fmt.Println(user)

	return user, nil
}

func Logout(token string) error {
	err := session.DestroySession(token)
	if err != nil {
		logrus.WithError(err).Error("failed to destroy user logged in token")
		return err
	}
	return nil
}

func ForgotPassword(ctx context.Context, email string) (*models.User, error) {
	user, err := mongod.GetUserByEmail(ctx, email)
	if err != nil {
		logrus.WithError(err).Error("failed to get user by email for password reset request")
		return nil, ErrFailedToGetUserByEmail
	}

	token, err := session.CreateSession(session.Session{
		AccountId: user.UserId,
		Validity: 15,
		UnitOfValidity: session.UnitOfValidityMinute,
	})

	if err != nil {
		logrus.WithError(err).Error("failed to create session for forgot password request")
		return nil, err
	}

	user.ResetPasswordToken = &token

	return user, nil
}

func ListUsers(ctx context.Context, filters models.ListUserFilter) (*models.UserList, error) {
	users, err := mongod.ListUsers(ctx, filters)
	if err != nil {
		logrus.WithError(err).Error(ErrListUsersFailed.Error())
		return nil, err
	}

	return users, nil
}

func ResetPassword(ctx context.Context, token, password, confirmPassword string) (*models.User, error) {
	if password != confirmPassword {
		logrus.WithError(ErrPasswordDoesNotMatch).Error(ErrPasswordDoesNotMatch)
		return nil, ErrPasswordDoesNotMatch
	}

	userSession, err := session.GetSessionByToken(token)
	if err != nil {
		logrus.WithError(err).Error("failed to valid session")
		return nil, ErrFailedToResetPasswordBadToken
	}

	user, err := mongod.GetUserById(ctx, userSession.AccountId)
	if err != nil {
		logrus.WithError(err).Error("failed to get user from database after validating user session")
		return nil, ErrFailedToResetPasswordBadToken
	}

	samePassword, _ := utils.VerifyPassword(user.Password, password)
	if samePassword {
		return nil, ErrPasswordIsSame
	}

	updatedUser, err := mongod.UpdateUser(ctx, &models.User{
		UserId: user.UserId,
		Password: utils.HashPassword(password),
	})
	if err != nil {
		logrus.WithError(err).Error("failed to update user password in database")
		return nil, ErrFailedToResetPassword
	}

	_ = session.DestroySession(token) // Destroy token once the password is reset

	return updatedUser, nil
}
