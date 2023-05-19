package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
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

func GetSessionByToken(token string) (*Session, error) {
	var ctx context.Context
	if strings.TrimSpace(token) == "" {
		return nil, ErrTokenInvalid
	}

	// verify token
	_, err := verifyAuthToken(token)
	if err != nil {
		logrus.WithError(err).Error("failed to confirm session validity")
		return nil, err
	}

	if err := mongod.SessionCollection().FindOne(ctx, bson.M{"token": token}); err != nil {
		logrus.WithError(err.Err()).Error("failed")
		return nil, ErrTokenSessionNotFound
	}

	var session Session

	if err = session.AssertValidity(); err != nil {
		logrus.WithError(err).Error("failed to get assert session validity")
		_ = DestroySession(session.Token) // Destroy it.
		return nil, err
	}

	return &session, nil
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
	var ctx context.Context

	if !payload.UnitOfValidity.IsValid() {
		return "", ErrInvalidUnitOfValidity
	}

	s := newSession(payload.AccountId, payload.Role, payload.Validity, payload.UnitOfValidity)
	_, err := mongod.SessionCollection().InsertOne(ctx, s)
	if err != nil {
		logrus.WithError(err).Error("failed to store session on db")
		fmt.Println("Error: ", err)
		return "", err
	}

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
