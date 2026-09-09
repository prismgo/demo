package repositories

import (
	"context"

	"gorm.io/gorm"

	"prismgo-demo/app/models"
)

// ReceiptRepositoryKey identifies the receipt repository in the service container.
const ReceiptRepositoryKey = "demo.repository.receipts"

// ReceiptRepository persists generated receipt metadata.
type ReceiptRepository struct{ db *gorm.DB }

// NewReceiptRepository creates a receipt repository backed by db.
func NewReceiptRepository(db *gorm.DB) *ReceiptRepository { return &ReceiptRepository{db: db} }

// Save persists receipt metadata.
func (r *ReceiptRepository) Save(ctx context.Context, receipt *models.Receipt) error {
	return r.db.WithContext(ctx).Save(receipt).Error
}

// FindByOrder returns receipt metadata for an order.
func (r *ReceiptRepository) FindByOrder(ctx context.Context, orderID uint) (*models.Receipt, error) {
	var receipt models.Receipt
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&receipt).Error; err != nil {
		return nil, err
	}
	return &receipt, nil
}
