package listeners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"order/src/application/commands"

	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	common_service "github.com/JohnSalazar/microservices-go-common/services"
	trace "github.com/JohnSalazar/microservices-go-common/trace/otel"
	"github.com/nats-io/nats.go"
)

type OrderCreateCommandListener struct {
	commandHandler *commands.OrderCommandHandler
	email          common_service.EmailService
	errorHelper    *common_nats.CommandErrorHelper
}

func NewOrderCreateCommandListener(
	commandHandler *commands.OrderCommandHandler,
	email common_service.EmailService,
	errorHelper *common_nats.CommandErrorHelper,
) *OrderCreateCommandListener {
	return &OrderCreateCommandListener{
		commandHandler: commandHandler,
		email:          email,
		errorHelper:    errorHelper,
	}
}

func (c *OrderCreateCommandListener) ProcessOrderCreateCommand() nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		_, span := trace.NewSpan(ctx, fmt.Sprintf("publish.%s\n", msg.Subject))
		defer span.End()

		orderCommand := &commands.CreateOrderCommand{}
		err := json.Unmarshal(msg.Data, orderCommand)
		if c.errorHelper.CheckUnmarshal(msg, err) == nil {
			err = c.commandHandler.CreateOrderCommandHandler(ctx, orderCommand)
			c.errorHelper.CheckCommandError(span, msg, err)
		}

		err = msg.Ack()
		if err != nil {
			log.Printf("stan msg.Ack error: %v\n", err)
		}
	}
}
