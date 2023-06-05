package dtos

import "github.com/google/uuid"

type UpdateStoreOrder struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"productid"`
}
