package models

import "time"

const (
	OrderStatusPending = "pending"
	OrderStatusPaid    = "paid"
)

// Order is the aggregate used by the documentation demo checkout flow.
type Order struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"index;not null"`
	Status     string     `json:"status" gorm:"size:32;index;not null"`
	TotalCents int64      `json:"total_cents" gorm:"not null"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// TableName keeps the demo schema explicit in examples and migrations.
func (Order) TableName() string { return "demo_orders" }
