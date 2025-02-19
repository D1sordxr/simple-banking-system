package transaction

import (
	"github.com/D1sordxr/simple-bank/bank-services/internal/domain/transaction"
	"github.com/D1sordxr/simple-bank/bank-services/internal/infrastructure/postgres/models"
	"github.com/google/uuid"
)

func ConvertAggregateToModel(agg transaction.Aggregate) models.TransactionModel {
	model := models.TransactionModel{
		ID:        agg.TransactionID.Value,
		Currency:  agg.Currency.String(),
		Amount:    agg.Amount.Value,
		Status:    agg.TransactionStatus.String(),
		Type:      agg.Type.String(),
		CreatedAt: agg.CreatedAt,
		UpdatedAt: agg.UpdatedAt,
	}

	if agg.SourceAccountID != nil && agg.SourceAccountID.Value != uuid.Nil {
		model.SourceAccountID = &agg.SourceAccountID.Value
	}

	if agg.DestinationAccountID != nil && agg.DestinationAccountID.Value != uuid.Nil {
		model.DestinationAccountID = &agg.DestinationAccountID.Value
	}

	if agg.Description != nil && agg.Description.Value != "" {
		model.Description = &agg.Description.Value
	}

	if agg.FailureReason != nil {
		model.FailureReason = agg.FailureReason
	}

	return model
}
