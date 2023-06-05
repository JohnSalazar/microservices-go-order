package commands

import (
	"order/src/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateStoreOrderCommand struct {
	ID     primitive.ObjectID `json:"id"`
	Stores []*models.Store    `json:"stores"`
}
