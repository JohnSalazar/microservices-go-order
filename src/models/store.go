package models

import "github.com/google/uuid"

type Store struct {
	ID        uuid.UUID `bson:"_id" json:"id"`
	ProductID uuid.UUID `bson:"product_id" json:"productid"`
}
