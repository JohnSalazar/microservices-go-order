package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	CustomerID primitive.ObjectID `bson:"customer_id" json:"customerId"`
	Products   []*Product         `bson:"products" json:"products"`
	Stores     []*Store           `bson:"stores" json:"stores"`
	Sum        float32            `bson:"sum" json:"sum"`
	Discount   float32            `bson:"discount" json:"discount"`
	Status     uint               `bson:"status" json:"status"`
	StatusAt   time.Time          `bson:"status_at" json:"status_at"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
	Version    uint               `bson:"version" json:"version"`
	Deleted    bool               `bson:"deleted" json:"deleted,omitempty"`
}
