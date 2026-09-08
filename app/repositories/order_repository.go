package repositories

import (
	"context"
	"errors"
	"time"

	"prismgo-demo/app/models"

	"gorm.io/gorm"
)

const OrderRepositoryKey = "demo.repository.orders"

var ErrOrderAlreadyPaid = errors.New("demo order is already paid")

// OrderRepository owns persistence operations for the order aggregate.
type OrderRepository struct{ db *gorm.DB }

func NewOrderRepository(db *gorm.DB) *OrderRepository { return &OrderRepository{db: db} }

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepository) Find(ctx context.Context, id uint) (*models.Order, error) {
	return r.find(r.db.WithContext(ctx), id)
}

// MarkPaid updates and returns an order inside the caller's transaction.
func (r *OrderRepository) MarkPaid(tx *gorm.DB, id uint, paidAt time.Time) (*models.Order, error) {
	order, err := r.find(tx, id)
	if err != nil {
		return nil, err
	}
	if order.Status == models.OrderStatusPaid {
		return nil, ErrOrderAlreadyPaid
	}
	if err := tx.Model(order).Updates(map[string]any{
		"status":  models.OrderStatusPaid,
		"paid_at": paidAt,
	}).Error; err != nil {
		return nil, err
	}
	order.Status = models.OrderStatusPaid
	order.PaidAt = &paidAt
	return order, nil
}

func (r *OrderRepository) DB() *gorm.DB { return r.db }

func (r *OrderRepository) find(db *gorm.DB, id uint) (*models.Order, error) {
	var order models.Order
	if err := db.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
