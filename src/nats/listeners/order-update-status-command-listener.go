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

type OrderUpdateStatusCommandListener struct {
	commandHandler *commands.OrderCommandHandler
	email          common_service.EmailService
	errorHelper    *common_nats.CommandErrorHelper
}

func NewOrderUpdateStatusCommandListener(
	commandHandler *commands.OrderCommandHandler,
	email common_service.EmailService,
	errorHelper *common_nats.CommandErrorHelper,
) *OrderUpdateStatusCommandListener {
	return &OrderUpdateStatusCommandListener{
		commandHandler: commandHandler,
		email:          email,
		errorHelper:    errorHelper,
	}
}

func (c *OrderUpdateStatusCommandListener) ProcessOrderUpdateStatusCommand() nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		_, span := trace.NewSpan(ctx, fmt.Sprintf("publish.%s\n", msg.Subject))
		defer span.End()

		orderCommand := &commands.UpdateStatusOrderCommand{}
		err := json.Unmarshal(msg.Data, orderCommand)
		if c.errorHelper.CheckUnmarshal(msg, err) == nil {
			err = c.commandHandler.UpdateStatusOrderCommandHandler(ctx, orderCommand)
			c.errorHelper.CheckCommandError(span, msg, err)
		}

		err = msg.Ack()
		if err != nil {
			log.Printf("stan msg.Ack error: %v\n", err)
		}
	}
}
