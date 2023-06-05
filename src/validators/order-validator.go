package validators

import (
	"order/src/dtos"
	"order/src/models"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	common_validator "github.com/oceano-dev/microservices-go-common/validators"
)

type addOrder struct {
	ID         primitive.ObjectID `from:"id" json:"id" validate:"required"`
	CustomerID primitive.ObjectID `from:"customerId" json:"customerId" validate:"required"`
	Products   []*models.Product  `from:"products" json:"products" validate:"required"`
	Sum        float32            `from:"sum" json:"sum" validate:"required"`
	Status     uint               `from:"status" json:"status"`
}

type updateStatusOrder struct {
	ID       primitive.ObjectID `from:"id" json:"id" validate:"required"`
	Status   uint               `from:"status" json:"status"`
	StatusAt time.Time          `from:"status_at" json:"status_at" validate:"required"`
}

type updateStoreOrder struct {
	ID        uuid.UUID `from:"id" json:"id" validate:"required"`
	ProductID uuid.UUID `from:"productId" json:"productId" validate:"required"`
}

func ValidateAddOrder(fields *dtos.AddOrder) interface{} {
	addOrder := addOrder{
		ID:         fields.ID,
		CustomerID: fields.CustomerID,
		Products:   fields.Products,
		Sum:        fields.Sum,
		Status:     fields.Status,
	}

	err := common_validator.Validate(addOrder)
	if err != nil {
		return err
	}

	return nil
}

func ValidateUpdateStatusOrder(fields *dtos.UpdateStatusOrder) interface{} {
	updateStatusOrder := updateStatusOrder{
		ID:       fields.ID,
		Status:   fields.Status,
		StatusAt: fields.StatusAt,
	}

	err := common_validator.Validate(updateStatusOrder)
	if err != nil {
		return err
	}

	return nil
}

func ValidateUpdateStoreOrder(fields *dtos.UpdateStoreOrder) interface{} {
	updateStoreOrder := updateStoreOrder{
		ID:        fields.ID,
		ProductID: fields.ProductID,
	}

	err := common_validator.Validate(updateStoreOrder)
	if err != nil {
		return err
	}

	return nil
}
