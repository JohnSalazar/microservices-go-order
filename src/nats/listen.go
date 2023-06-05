package nats

import (
	"order/src/application/commands"
	"order/src/nats/listeners"

	"github.com/nats-io/nats.go"
	"github.com/oceano-dev/microservices-go-common/config"

	common_nats "github.com/oceano-dev/microservices-go-common/nats"
	common_service "github.com/oceano-dev/microservices-go-common/services"
)

type listen struct {
	js nats.JetStreamContext
}

const queueGroupName string = "orders-service"

var (
	subscribe          common_nats.Listener
	commandErrorHelper *common_nats.CommandErrorHelper

	orderCreateCommand       *listeners.OrderCreateCommandListener
	orderUpdateStatusCommand *listeners.OrderUpdateStatusCommandListener
	orderUpdateStoreCommand  *listeners.OrderUpdateStoreCommandListener
)

func NewListen(
	config *config.Config,
	js nats.JetStreamContext,
	orderCommandHandler *commands.OrderCommandHandler,
	email common_service.EmailService,
) *listen {
	subscribe = common_nats.NewListener(js)
	commandErrorHelper = common_nats.NewCommandErrorHelper(config, email)

	orderCreateCommand = listeners.NewOrderCreateCommandListener(orderCommandHandler, email, commandErrorHelper)
	orderUpdateStatusCommand = listeners.NewOrderUpdateStatusCommandListener(orderCommandHandler, email, commandErrorHelper)
	orderUpdateStoreCommand = listeners.NewOrderUpdateStoreCommandListener(orderCommandHandler, email, commandErrorHelper)
	return &listen{
		js: js,
	}
}

func (l *listen) Listen() {
	go subscribe.Listener(string(common_nats.OrderCreate), queueGroupName, queueGroupName+"_0", orderCreateCommand.ProcessOrderCreateCommand())
	go subscribe.Listener(string(common_nats.OrderStatus), queueGroupName, queueGroupName+"_1", orderUpdateStatusCommand.ProcessOrderUpdateStatusCommand())
	go subscribe.Listener(string(common_nats.StoreBooked), queueGroupName, queueGroupName+"_2", orderUpdateStoreCommand.ProcessOrderUpdateStoreCommand())
}
