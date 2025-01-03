package handlers

import (
	"context"
	"github.com/D1sordxr/simple-banking-system/internal/application/persistence"
	"github.com/D1sordxr/simple-banking-system/internal/application/transaction/commands"
	"github.com/D1sordxr/simple-banking-system/internal/domain/shared/event"
	"github.com/D1sordxr/simple-banking-system/internal/domain/shared/outbox"
	sharedVO "github.com/D1sordxr/simple-banking-system/internal/domain/shared/shared_vo"
	"github.com/D1sordxr/simple-banking-system/internal/domain/transaction"
	"github.com/D1sordxr/simple-banking-system/internal/domain/transaction/vo"
)

type CreateTransactionHandler struct {
	Repository transaction.Repository
	UoWManager persistence.UoWManager
}

func NewCreateTransactionHandler(repo transaction.Repository,
	uow persistence.UoWManager) *CreateTransactionHandler {
	return &CreateTransactionHandler{
		Repository: repo,
		UoWManager: uow,
	}
}

func (h CreateTransactionHandler) Handle(ctx context.Context,
	c commands.CreateTransactionCommand) (commands.CreateTransactionDTO, error) {

	txID := sharedVO.NewUUID()

	var err error
	var sourceAccountID, destinationAccountID *sharedVO.UUID
	if len(c.SourceAccountID) != 0 {
		sourceAccountID, err = sharedVO.NewPointerUUIDFromString(c.SourceAccountID)
		if err != nil {
			return commands.CreateTransactionDTO{}, err
		}
	}
	if len(c.DestinationAccountID) != 0 {
		destinationAccountID, err = sharedVO.NewPointerUUIDFromString(c.DestinationAccountID)
		if err != nil {
			return commands.CreateTransactionDTO{}, err
		}
	}

	currency, err := sharedVO.NewCurrency(c.Currency)
	if err != nil {
		return commands.CreateTransactionDTO{}, err
	}
	amount := sharedVO.NewMoneyFromFloat(c.Amount)
	txStatus := vo.NewTransactionStatus()
	txType, err := vo.NewType(c.Type)
	if err != nil {
		return commands.CreateTransactionDTO{}, err
	}
	description, err := vo.NewDescription(c.Description)
	if err != nil {
		return commands.CreateTransactionDTO{}, err
	}

	txAggregate, err := transaction.NewTransaction(
		txID, sourceAccountID, destinationAccountID, currency, amount, txStatus, txType, description)
	if err != nil {
		return commands.CreateTransactionDTO{}, err
	}

	uow := h.UoWManager.GetUoW()
	tx, err := uow.Begin(ctx)
	if err != nil {
		return commands.CreateTransactionDTO{}, err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = uow.Rollback(ctx)
			panic(r)
		}
		if err != nil {
			_ = uow.Rollback(ctx)
		}
	}()

	if err = h.Repository.Create(ctx, tx, txAggregate); err != nil {
		return commands.CreateTransactionDTO{}, err
	}

	txEvent, err := event.NewTransactionCreatedEvent(txAggregate)
	if err != nil {
		return commands.CreateTransactionDTO{}, err
	}
	if err = h.Repository.SaveEvent(ctx, tx, txEvent); err != nil {
		return commands.CreateTransactionDTO{}, err
	}

	outboxEvent, err := outbox.NewOutboxEvent(txEvent)
	if err != nil {
		return commands.CreateTransactionDTO{}, err
	}
	if err = h.Repository.SaveOutboxEvent(ctx, tx, outboxEvent); err != nil {
		return commands.CreateTransactionDTO{}, err
	}

	if err = uow.Commit(ctx); err != nil {
		return commands.CreateTransactionDTO{}, err
	}

	return commands.CreateTransactionDTO{
		TransactionID: txID.String(),
	}, nil
}
