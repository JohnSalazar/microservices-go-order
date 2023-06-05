package events

import (
	"order/src/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderStatusUpdatedEvent struct {
	ID        primitive.ObjectID `json:"id"`
	Products  []*models.Product  `json:"products"`
	Stores    []*models.Store    `json:"stores"`
	Status    uint               `json:"status"`
	StatusAt  time.Time          `json:"status_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Version   uint               `json:"version"`
}
