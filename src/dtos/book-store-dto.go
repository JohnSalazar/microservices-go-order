package dtos

import (
	"order/src/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookStore struct {
	OrderID  primitive.ObjectID `json:"orderId"`
	Products []*models.Product  `json:"products"`
}
