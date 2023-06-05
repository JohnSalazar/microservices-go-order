package events

import (
	"context"
	"encoding/json"
	"fmt"
	"order/src/dtos"
	"time"

	common_models "github.com/oceano-dev/microservices-go-common/models"
	common_nats "github.com/oceano-dev/microservices-go-common/nats"
	common_service "github.com/oceano-dev/microservices-go-common/services"
)

type OrderEventHandler struct {
	email     common_service.EmailService
	publisher common_nats.Publisher
}

func NewOrderEventHandler(
	email common_service.EmailService,
	publisher common_nats.Publisher,
) *OrderEventHandler {
	return &OrderEventHandler{
		email:     email,
		publisher: publisher,
	}
}

func (order *OrderEventHandler) OrderCreatedEventHandler(ctx context.Context, event *OrderCreatedEvent) error {

	payment := map[string]interface{}{
		"orderId":    event.ID,
		"total":      event.Sum - event.Discount,
		"cardNumber": event.CardNumber,
		"kid":        event.Kid,
	}

	dataPayment, _ := json.Marshal(payment)
	err := order.publisher.Publish(string(common_nats.PaymentCreate), dataPayment)
	if err != nil {
		return err
	}

	cart := map[string]interface{}{
		"id": event.ID,
	}

	dataCart, _ := json.Marshal(cart)
	err = order.publisher.Publish(string(common_nats.OrderCreated), dataCart)
	if err != nil {
		return err
	}

	return nil
}

func (order *OrderEventHandler) OrderStatusUpdatedEventHandler(ctx context.Context, event *OrderStatusUpdatedEvent) error {
	// fmt.Println(event)

	if event.Status == uint(common_models.OrderCanceled) {

		updateStatusPaymentByOrder := &dtos.UpdateStatusPaymentByOrder{
			OrderID:  event.ID,
			Status:   uint(common_models.PaymentCanceled),
			StatusAt: time.Now().UTC(),
		}

		dataPayment, _ := json.Marshal(updateStatusPaymentByOrder)
		err := order.publisher.Publish(string(common_nats.PaymentCancel), dataPayment)
		if err != nil {
			return err
		}

		go order.email.SendSupportMessage(fmt.Sprintf("Order ID: %s canceled", event.ID))
	}

	if event.Status == uint(common_models.PaymentConfirmed) {
		bookStoreDto := &dtos.BookStore{
			OrderID:  event.ID,
			Products: event.Products,
		}

		data, _ := json.Marshal(bookStoreDto)
		err := order.publisher.Publish(string(common_nats.StoreBook), data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (order *OrderEventHandler) OrderStoreUpdatedEventHandler(ctx context.Context, event *OrderStoreUpdatedEvent) error {

	paymentStoreCommand := map[string]interface{}{
		"orderId": event.ID,
		"stores":  event.Stores,
	}

	data, _ := json.Marshal(paymentStoreCommand)
	err := order.publisher.Publish(string(common_nats.StorePayment), data)
	if err != nil {
		return err
	}

	return nil
}
