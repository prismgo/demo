// Package repositories contains the demo application's persistence adapters.
package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"prismgo-demo/app/models"
)

// OrderRepositoryKey identifies the order repository in the service container.
const OrderRepositoryKey = "demo.repository.orders"

// ErrOrderAlreadyPaid reports an attempt to pay an already-paid order.
var ErrOrderAlreadyPaid = errors.New("demo order is already paid")

// OrderRepository owns persistence operations for the order aggregate.
type OrderRepository struct{ db *gorm.DB }

// NewOrderRepository creates an order repository backed by db.
func NewOrderRepository(db *gorm.DB) *OrderRepository { return &OrderRepository{db: db} }

// Create persists a new order.
func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// Find returns an order by ID.
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

// DB returns the underlying database handle for transaction orchestration.
func (r *OrderRepository) DB() *gorm.DB { return r.db }

func (r *OrderRepository) find(db *gorm.DB, id uint) (*models.Order, error) {
	var order models.Order
	if err := db.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
