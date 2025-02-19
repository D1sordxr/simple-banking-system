package handlers

import (
	"context"
	"fmt"
	"github.com/D1sordxr/simple-bank/bank-services/internal/application/account/commands"
	"github.com/D1sordxr/simple-bank/bank-services/internal/application/account/dependencies"
	accountRoot "github.com/D1sordxr/simple-bank/bank-services/internal/domain/account"
	"github.com/D1sordxr/simple-bank/bank-services/internal/domain/account/vo"
	"github.com/D1sordxr/simple-bank/bank-services/internal/domain/shared/event"
	"github.com/D1sordxr/simple-bank/bank-services/internal/domain/shared/outbox"
	sharedExceptions "github.com/D1sordxr/simple-bank/bank-services/internal/domain/shared/shared_exceptions"
	sharedVO "github.com/D1sordxr/simple-bank/bank-services/internal/domain/shared/shared_vo"
)

type CreateAccountHandler struct {
	deps *dependencies.Dependencies
}

func NewCreateAccountHandler(dependencies *dependencies.Dependencies) *CreateAccountHandler {
	return &CreateAccountHandler{
		deps: dependencies,
	}
}

func (h *CreateAccountHandler) Handle(ctx context.Context, c commands.CreateAccountCommand) (commands.CreateDTO, error) {
	const op = "Services.AccountService.CreateAccount"

	logger := h.deps.Logger
	log := logger.With(
		logger.String("operation", op),
		logger.Group("account",
			logger.String("clientID", c.ClientID),
			logger.String("currency", c.Currency),
		),
	)

	log.Info("Attempting to create new account")

	clientID, err := sharedVO.NewUUIDFromString(c.ClientID)
	if err != nil {
		log.Error(sharedExceptions.LogVOCreationError("UUID"), sharedExceptions.LogError(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	accountID := sharedVO.NewUUID()

	balance := vo.NewBalance()

	currency, err := sharedVO.NewCurrency(c.Currency)
	if err != nil {
		log.Error(sharedExceptions.LogVOCreationError("currency"), sharedExceptions.LogError(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	status := vo.NewStatus()

	account, err := accountRoot.NewAccount(accountID, clientID, balance, currency, status)
	if err != nil {
		log.Error(sharedExceptions.LogAggregateCreationError("account"), sharedExceptions.LogError(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	uow := h.deps.UnitOfWork
	ctx, err = uow.BeginWithTx(ctx)
	if err != nil {
		log.Error(sharedExceptions.LogErrorAsString(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	defer uow.GracefulRollback(ctx, &err)

	if err = h.deps.AccountRepository.Create(ctx, account); err != nil {
		log.Error(sharedExceptions.LogErrorAsString(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	accountEvent, err := event.NewAccountCreatedEvent(account)
	if err != nil {
		log.Error(sharedExceptions.LogEventCreationError(), sharedExceptions.LogError(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if err = h.deps.EventRepository.SaveEvent(ctx, accountEvent); err != nil {
		log.Error(sharedExceptions.LogErrorAsString(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	outboxEvent, err := outbox.NewOutboxEvent(accountEvent)
	if err != nil {
		log.Error(sharedExceptions.LogOutboxCreationError(), sharedExceptions.LogError(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}
	if err = h.deps.OutboxRepository.SaveOutboxEvent(ctx, outboxEvent); err != nil {
		log.Error(sharedExceptions.LogErrorAsString(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	if err = uow.Commit(ctx); err != nil {
		log.Error(sharedExceptions.LogErrorAsString(err))
		return commands.CreateDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Account creation completed successfully")
	return commands.CreateDTO{
		AccountID: accountID.String(),
	}, nil
}
