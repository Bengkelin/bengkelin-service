package dto

import (
	"github.com/Bengkelin/bengkelin-service/internal/pkg/models"
)

type BengkelDto struct {
	ID           string                      `json:"bengkel_id"`
	BengkelName  string                      `json:"bengkel_name"`
	BengkelPhoto string                      `json:"bengkel_photo"`
	Address      models.BengkelAddress       `json:"address"`
	Distance     float64                     `json:"distance"`
	Operasionals []models.BengkelOperasional `json:"operasionals"`
}
