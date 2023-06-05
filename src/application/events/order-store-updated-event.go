package events

import (
	"order/src/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderStoreUpdatedEvent struct {
	ID        primitive.ObjectID `json:"id"`
	Stores    []*models.Store    `json:"stores"`
	UpdatedAt time.Time          `json:"updated_at"`
	Version   uint               `json:"version"`
}
