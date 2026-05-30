package helpers

import (
	"math"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	uuid := uuid.New().String()
	return uuid
}

func CalculateDistanceHaversineAlg(lat1, lon1, lat2, lon2 float64) float64 {
	const radius = 6371 // Earth radius in km
	dLat := degToRad(lat2 - lat1)
	dLon := degToRad(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(degToRad(lat1))*math.Cos(degToRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := radius * c

	return distance
}

func degToRad(deg float64) float64 {
	return deg * (math.Pi / 180)
}
