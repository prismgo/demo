package repositories

import (
	"context"

	"prismgo-demo/app/models"

	"gorm.io/gorm"
)

const ReceiptRepositoryKey = "demo.repository.receipts"

// ReceiptRepository persists generated receipt metadata.
type ReceiptRepository struct{ db *gorm.DB }

func NewReceiptRepository(db *gorm.DB) *ReceiptRepository { return &ReceiptRepository{db: db} }

func (r *ReceiptRepository) Save(ctx context.Context, receipt *models.Receipt) error {
	return r.db.WithContext(ctx).Save(receipt).Error
}

func (r *ReceiptRepository) FindByOrder(ctx context.Context, orderID uint) (*models.Receipt, error) {
	var receipt models.Receipt
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&receipt).Error; err != nil {
		return nil, err
	}
	return &receipt, nil
}
