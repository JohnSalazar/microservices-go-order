package interfaces

import (
	"context"

	"order/src/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderRepository interface {
	GetAll(ctx context.Context, customerID primitive.ObjectID) ([]*models.Order, error)
	FindByCustomerID(ctx context.Context, customerID primitive.ObjectID) (*models.Order, error)
	FindByID(ctx context.Context, ID primitive.ObjectID) (*models.Order, error)
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) (*models.Order, error)
	Delete(ctx context.Context, ID primitive.ObjectID) error
}
