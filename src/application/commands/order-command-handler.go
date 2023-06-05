package commands

import (
	"context"
	"errors"
	"order/src/application/events"
	"order/src/dtos"
	"order/src/models"
	"order/src/repositories/interfaces"
	"order/src/validators"
	"strings"
	"time"

	common_models "github.com/oceano-dev/microservices-go-common/models"
)

type OrderCommandHandler struct {
	orderRepository   interfaces.OrderRepository
	orderEventHandler *events.OrderEventHandler
}

func NewOrderCommandHandler(
	orderRepository interfaces.OrderRepository,
	orderEventHandler *events.OrderEventHandler,
) *OrderCommandHandler {
	return &OrderCommandHandler{
		orderRepository:   orderRepository,
		orderEventHandler: orderEventHandler,
	}
}

func (order *OrderCommandHandler) CreateOrderCommandHandler(ctx context.Context, command *CreateOrderCommand) error {

	orderDto := &dtos.AddOrder{
		ID:         command.ID,
		CustomerID: command.CustomerID,
		Products:   command.Products,
		Sum:        command.Sum,
		Discount:   command.Discount,
		Status:     uint(common_models.OrderCreated),
	}

	result := validators.ValidateAddOrder(orderDto)
	if result != nil {
		return errors.New(strings.Join(result.([]string), ""))
	}

	orderModel := &models.Order{
		ID:         orderDto.ID,
		CustomerID: orderDto.CustomerID,
		Products:   orderDto.Products,
		Sum:        orderDto.Sum,
		Discount:   orderDto.Discount,
		Status:     orderDto.Status,
		StatusAt:   time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
	}

	orderExists, _ := order.orderRepository.FindByID(ctx, orderDto.ID)
	if orderExists != nil {
		return errors.New("already a order for this customer")
	}

	orderModel, err := order.orderRepository.Create(ctx, orderModel)
	if err != nil {
		return err
	}

	orderEvent := &events.OrderCreatedEvent{
		ID:         orderModel.ID,
		CustomerID: orderModel.CustomerID,
		Products:   orderModel.Products,
		Sum:        orderModel.Sum,
		Discount:   orderModel.Discount,
		Status:     orderModel.Status,
		StatusAt:   orderModel.StatusAt,
		CardNumber: command.CardNumber,
		Kid:        command.Kid,
		CreatedAt:  orderModel.CreatedAt,
		Version:    orderModel.Version,
	}

	go order.orderEventHandler.OrderCreatedEventHandler(ctx, orderEvent)

	return nil
}

func (order *OrderCommandHandler) UpdateStatusOrderCommandHandler(ctx context.Context, command *UpdateStatusOrderCommand) error {
	orderDto := *&dtos.UpdateStatusOrder{
		ID:       command.ID,
		Status:   command.Status,
		StatusAt: command.StatusAt,
	}

	result := validators.ValidateUpdateStatusOrder(&orderDto)
	if result != nil {
		return errors.New(strings.Join(result.([]string), ""))
	}

	orderExists, err := order.orderRepository.FindByID(ctx, orderDto.ID)
	if err != nil {
		return err
	}

	orderModel := &models.Order{
		ID:        orderDto.ID,
		Products:  orderExists.Products,
		Stores:    orderExists.Stores,
		Sum:       orderExists.Sum,
		Discount:  orderExists.Discount,
		Status:    orderDto.Status,
		StatusAt:  orderDto.StatusAt,
		UpdatedAt: time.Now().UTC(),
		Version:   orderExists.Version,
	}

	orderModel, err = order.orderRepository.Update(ctx, orderModel)
	if err != nil {
		return err
	}

	orderEvent := &events.OrderStatusUpdatedEvent{
		ID:        orderModel.ID,
		Products:  orderModel.Products,
		Stores:    orderModel.Stores,
		Status:    orderModel.Status,
		StatusAt:  orderModel.StatusAt,
		UpdatedAt: orderModel.UpdatedAt,
		Version:   orderModel.Version,
	}

	go order.orderEventHandler.OrderStatusUpdatedEventHandler(ctx, orderEvent)

	return nil
}

func (order *OrderCommandHandler) UpdateStoreOrderCommandHandler(ctx context.Context, command *UpdateStoreOrderCommand) error {
	stores := []*models.Store{}
	for _, store := range command.Stores {
		orderDto := &dtos.UpdateStoreOrder{
			ID:        store.ID,
			ProductID: store.ProductID,
		}

		result := validators.ValidateUpdateStoreOrder(orderDto)
		if result != nil {
			return errors.New(strings.Join(result.([]string), ""))
		}

		storeModel := &models.Store{
			ID:        store.ID,
			ProductID: store.ProductID,
		}

		stores = append(stores, storeModel)
	}

	orderExists, err := order.orderRepository.FindByID(ctx, command.ID)
	if err != nil {
		return err
	}

	orderModel := &models.Order{
		ID:         orderExists.ID,
		CustomerID: orderExists.CustomerID,
		Products:   orderExists.Products,
		Stores:     stores,
		Sum:        orderExists.Sum,
		Discount:   orderExists.Discount,
		Status:     orderExists.Status,
		StatusAt:   orderExists.StatusAt,
		UpdatedAt:  time.Now().UTC(),
		Version:    orderExists.Version,
	}

	orderModel, err = order.orderRepository.Update(ctx, orderModel)
	if err != nil {
		return err
	}

	orderEvent := &events.OrderStoreUpdatedEvent{
		ID:        orderModel.ID,
		Stores:    stores,
		UpdatedAt: orderModel.UpdatedAt,
		Version:   orderModel.Version,
	}

	go order.orderEventHandler.OrderStoreUpdatedEventHandler(ctx, orderEvent)

	return nil
}
