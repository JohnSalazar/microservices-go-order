package commands

import (
	"order/src/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateOrderCommand struct {
	ID         primitive.ObjectID `json:"id"`
	CustomerID primitive.ObjectID `json:"customerId"`
	Products   []*models.Product  `json:"products"`
	Stores     []*models.Store    `json:"stores"`
	Sum        float32            `json:"sum"`
	Discount   float32            `json:"discount"`
	CardNumber []byte             `json:"cardNumber"`
	Kid        string             `json:"kid"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at,omitempty"`
	Version    uint               `json:"version"`
	Deleted    bool               `json:"deleted,omitempty"`
}
