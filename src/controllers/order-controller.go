package controllers

import (
	"net/http"

	"order/src/repositories/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/oceano-dev/microservices-go-common/helpers"
	"github.com/oceano-dev/microservices-go-common/httputil"
	trace "github.com/oceano-dev/microservices-go-common/trace/otel"
)

type OrderController struct {
	orderRepository interfaces.OrderRepository
}

func NewOrderController(
	orderRepository interfaces.OrderRepository,
) *OrderController {
	return &OrderController{
		orderRepository: orderRepository,
	}
}

func (order *OrderController) GetAll(c *gin.Context) {
	_, span := trace.NewSpan(c.Request.Context(), "OrderController.GetAll")
	defer span.End()

	ID, customerIDOk := c.Get("user")
	if !customerIDOk {
		httputil.NewResponseError(c, http.StatusForbidden, "invalid customer")
		return
	}

	isID := helpers.IsValidID(ID.(string))
	if !isID {
		httputil.NewResponseError(c, http.StatusBadRequest, "invalid customerId")
		return
	}

	customerID := helpers.StringToID(ID.(string))

	orders, err := order.orderRepository.GetAll(c.Request.Context(), customerID)
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "orders get error")
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (order *OrderController) GetById(c *gin.Context) {
	_, span := trace.NewSpan(c.Request.Context(), "OrderController.GetById")
	defer span.End()

	ID, customerIDOk := c.Get("user")
	if !customerIDOk {
		httputil.NewResponseError(c, http.StatusForbidden, "invalid customer")
		return
	}

	isID := helpers.IsValidID(ID.(string))
	if !isID {
		httputil.NewResponseError(c, http.StatusBadRequest, "invalid customerId")
		return
	}

	customerID := helpers.StringToID(ID.(string))

	orderModel, err := order.orderRepository.FindByCustomerID(c.Request.Context(), customerID)
	if orderModel == nil || err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "order not found")
		return
	}

	c.JSON(http.StatusOK, orderModel)
}
