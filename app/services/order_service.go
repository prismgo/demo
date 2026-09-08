package services

import (
	"context"
	"errors"
	"time"

	"prismgo-demo/app/models"
	"prismgo-demo/app/repositories"

	"gorm.io/gorm"
)

const OrderServiceKey = "demo.service.orders"

var ErrInvalidOrderTotal = errors.New("demo order total must be positive")

// OrderService coordinates the transactional part of the fulfillment flow.
// Cache, event, queue and filesystem collaboration is added by later batches.
type OrderService struct {
	orders *repositories.OrderRepository
	now    func() time.Time
}

func NewOrderService(orders *repositories.OrderRepository) *OrderService {
	return &OrderService{orders: orders, now: time.Now}
}

func (s *OrderService) Create(ctx context.Context, userID uint, totalCents int64) (*models.Order, error) {
	if totalCents <= 0 {
		return nil, ErrInvalidOrderTotal
	}
	order := &models.Order{UserID: userID, Status: models.OrderStatusPending, TotalCents: totalCents}
	if err := s.orders.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) Find(ctx context.Context, id uint) (*models.Order, error) {
	return s.orders.Find(ctx, id)
}

func (s *OrderService) Pay(ctx context.Context, id uint) (*models.Order, error) {
	var paid *models.Order
	err := s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := s.orders.MarkPaid(tx, id, s.now().UTC())
		if err != nil {
			return err
		}
		paid = order
		return nil
	})
	return paid, err
}
