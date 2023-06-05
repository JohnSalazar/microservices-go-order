package events

import (
	"order/src/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderCreatedEvent struct {
	ID         primitive.ObjectID `json:"id"`
	CustomerID primitive.ObjectID `json:"customerId"`
	Products   []*models.Product  `json:"products"`
	Sum        float32            `json:"sum"`
	Discount   float32            `json:"discount"`
	Status     uint               `json:"status"`
	StatusAt   time.Time          `json:"status_at"`
	CardNumber []byte             `json:"cardNumber"`
	Kid        string             `json:"kid"`
	CreatedAt  time.Time          `json:"created_at"`
	Version    uint               `json:"version"`
}
