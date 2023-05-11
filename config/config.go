package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Secrets struct {
	DatabaseURL  string `json:"DATABASE_URL"`
	DatabaseName string `json:"DATABASE_NAME"`
	Port         string 
	JWTSecret    string `json:"JWT_SECRET"`
}

var ss Secrets

func init() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading .env file")
	}

	ss = Secrets{}

	ss.DatabaseURL = os.Getenv("DATABASE_URL")
	ss.DatabaseName = os.Getenv("DATABASE_NAME")
	ss.JWTSecret = os.Getenv("JWT_SECRET")

	if ss.Port = os.Getenv("PORT"); ss.Port == "" {
		ss.Port = "80"
	}

}

func GetSecrets() Secrets {
	return ss
}
