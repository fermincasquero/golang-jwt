package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fredele20/golang-jwt-project/database/mongod"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrTokenInvalid          = errors.New("invalid token string provided")
	ErrTokenExpired          = errors.New("sorry, session has expired. Please login again to continue")
	ErrTokenSessionNotFound  = errors.New("session not found or destroyed")
	ErrInvalidUnitOfValidity = errors.New("invalid unit of validity, you must provide HOUR or MINUTE")
)

func generateToken(id, role string) string {
	payload := &TokenPayload{
		Role: role,
		Id:   id,
		Payload: jwt.Payload{
			Issuer:   "Golang",
			Subject:  "Golang JWT",
			Audience: jwt.Audience{""},
			IssuedAt: jwt.NumericDate(time.Now()),
			JWTID:    "Golang JWT Auth",
		},
	}

	token, err := jwt.Sign(payload, jwt.NewHS256([]byte(os.Getenv("JWT_SECRET"))))
	if err != nil {
		logrus.Debugf("error generating JWT Token: %s", err)
		return ""
	}

	return string(token)
}

func verifyAuthToken(token string) (*TokenPayload, error) {
	secret := jwt.NewHS256([]byte(os.Getenv("JWT_SECRET")))
	var payloadBody TokenPayload
	_, err := jwt.Verify([]byte(token), secret, &payloadBody)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	return &payloadBody, nil
}

func newSession(accountId, role string, validity time.Duration, unitOfValidity UnitOfValidity) *Session {
	token := generateToken(accountId, role)
	return &Session{
		Token:          token,
		Role:           role,
		AccountId:      accountId,
		Validity:       validity,
		LastUsage:      time.Now(),
		UnitOfValidity: unitOfValidity,
		TimeCreated:    time.Now(),
	}
}

func CreateSession(payload Session) (string, error) {
	fmt.Println("Function called")
	var ctx context.Context

	if !payload.UnitOfValidity.IsValid() {
		return "", ErrInvalidUnitOfValidity
	}

	// input := make(map[string][]byte)

	s := newSession(payload.AccountId, payload.Role, payload.Validity, payload.UnitOfValidity)
	_, err := mongod.SessionCollection().InsertOne(ctx, s)
	if err != nil {
		logrus.WithError(err).Error("failed to store session on db")
		fmt.Println("Error: ", err)
		return "", err
	}

	fmt.Println("Token: ", s.Token)

	return s.Token, nil
}

func DestroySession(token string) error {
	// Delete session from the DB
	var ctx context.Context
	session := mongod.SessionCollection().FindOneAndDelete(ctx, bson.M{"token":token})
	if session.Err() != nil {
		logrus.WithError(session.Err()).Error("session with the token not found")
		return session.Err()
	}

	return nil
}
