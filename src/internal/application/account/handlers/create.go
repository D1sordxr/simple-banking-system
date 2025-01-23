package handlers

import (
	"context"
	accountDeps "github.com/D1sordxr/simple-banking-system/internal/application/account"
	"github.com/D1sordxr/simple-banking-system/internal/application/account/commands"
	accountRoot "github.com/D1sordxr/simple-banking-system/internal/domain/account"
	"github.com/D1sordxr/simple-banking-system/internal/domain/account/vo"
	sharedExceptions "github.com/D1sordxr/simple-banking-system/internal/domain/shared/shared_exceptions"
	sharedVO "github.com/D1sordxr/simple-banking-system/internal/domain/shared/shared_vo"
	"log/slog"
)

type CreateAccountHandler struct {
	deps *accountDeps.Dependencies
}

func NewCreateAccountHandler(dependencies *accountDeps.Dependencies) *CreateAccountHandler {
	return &CreateAccountHandler{
		deps: dependencies,
	}
}

func (h *CreateAccountHandler) Handle(ctx context.Context, c commands.CreateAccountCommand) (commands.CreateDTO, error) {
	const op = "Services.AccountService.CreateAccount"

	log := h.deps.Logger.With(
		slog.String("operation", op),
		slog.String("clientID", c.ClientID),
	)

	log.Info("Attempting to create new account")

	clientID, err := sharedVO.NewUUIDFromString(c.ClientID)
	if err != nil {
		log.Error(sharedExceptions.LogVOCreationError("UUID"), sharedExceptions.LogError(err))
		return commands.CreateDTO{}, err
	}

	accountID := sharedVO.NewUUID()

	balance := vo.NewBalance()

	currency, err := sharedVO.NewCurrency(c.Currency)
	if err != nil {
		log.Error(sharedExceptions.LogVOCreationError("currency"), sharedExceptions.LogError(err))
		return commands.CreateDTO{}, err
	}

	status := vo.NewStatus()

	account, err := accountRoot.NewAccount(clientID, accountID, balance, currency, status)
	if err != nil {
		log.Error(sharedExceptions.LogAggregateCreationError("account"))
		return commands.CreateDTO{}, err
	}

	uow := h.deps.UoWManager.GetUoW()
	tx, err := uow.Begin()
	if err != nil {
		log.Error(sharedExceptions.LogErrorAsString(err))
		return commands.CreateDTO{}, err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = uow.Rollback()
			panic(r)
		}
		if err != nil {
			log.Error(sharedExceptions.LogErrorAsString(err))
			_ = uow.Rollback()
		}
	}()

	if err = h.deps.AccountRepository.Create(ctx, tx, account); err != nil {
		log.Error(sharedExceptions.LogErrorAsString(err))
		return commands.CreateDTO{}, err
	}

	// TODO: add event and outbox

	if err = uow.Commit(); err != nil {
		return commands.CreateDTO{}, err
	}

	return commands.CreateDTO{
		AccountID: accountID.String(),
	}, nil
}
