package helpers

import "github.com/google/uuid"

func GenerateUUID() string {
	uuid := uuid.New().String()
	return uuid
}