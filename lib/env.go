package lib

import (
	"github.com/joho/godotenv"
)

// LoadEnv loads env variables from .env file
func LoadEnv(filename string) error {
	err := godotenv.Load(filename)
	if err != nil {
		return err
	}

	return nil
}
