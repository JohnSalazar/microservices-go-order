package dtos

import (
	"order/src/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AddOrder struct {
	ID         primitive.ObjectID `json:"id"`
	CustomerID primitive.ObjectID `json:"customerId"`
	Products   []*models.Product  `json:"products"`
	Stores     []*models.Store    `json:"stores"`
	Sum        float32            `json:"sum"`
	Discount   float32            `json:"discount"`
	Status     uint               `json:"status"`
}
