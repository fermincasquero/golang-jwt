package session

import (
	"encoding/json"
	"time"

	"github.com/gbrlsnchs/jwt/v3"
)

type UnitOfValidity string

type Session struct {
	Token          string         `json:"token"`
	Role           string         `json:"role"`
	AccountId      string         `json:"accountId"`
	TimeCreated    time.Time      `json:"timeCreated"`
	Validity       time.Duration  `json:"validity"`
	UnitOfValidity UnitOfValidity `json:"unitOfValidity"`
	LastUsage      time.Time      `json:"lastUsage"`
}

type TokenPayload struct {
	Id   string `json:"id"`
	Role string `json:"client"`
	jwt.Payload
}

const (
	UnitOfValidityMinute UnitOfValidity = "MINUTE"
	UnitOfValidityHour   UnitOfValidity = "HOUR"
)

func (u UnitOfValidity) IsValid() bool {
	switch u {
	case UnitOfValidityHour, UnitOfValidityMinute:
		return true
	}
	return false
}

func (sm *Session) AssertValidity() error {
	now := time.Now().Unix()
	lastUsage := sm.LastUsage
	uV := sm.UnitOfValidity
	var validity int64 = 0
	switch uV {
	case UnitOfValidityHour:
		validity = lastUsage.Add(time.Hour * sm.Validity).Unix()
	case UnitOfValidityMinute:
		validity = lastUsage.Add(time.Minute * sm.Validity).Unix()
	}

	if now > validity {
		return ErrTokenExpired
	}
	return nil
}

func (u UnitOfValidity) String() string {
	return string(u)
}

func (sm *Session) Byte() []byte {
	if sm == nil {
		return nil
	}

	b, _ := json.Marshal(sm)

	return b
}
