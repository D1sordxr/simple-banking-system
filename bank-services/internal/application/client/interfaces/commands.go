package interfaces

import (
	"context"
	"github.com/D1sordxr/simple-bank/bank-services/internal/application/client/commands"
)

type CreateClientCommand interface {
	Handle(ctx context.Context, c commands.CreateClientCommand) (commands.CreateDTO, error)
}

type UpdateClientCommand interface {
	Handle(ctx context.Context, c commands.UpdateClientCommand) (commands.UpdateDTO, error)
}
