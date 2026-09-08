package models

import "time"

// Receipt records the generated filesystem artifact for a paid order.
type Receipt struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	OrderID     uint      `json:"order_id" gorm:"uniqueIndex;not null"`
	Path        string    `json:"path" gorm:"size:512;not null"`
	GeneratedAt time.Time `json:"generated_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName keeps receipt records isolated from host application tables.
func (Receipt) TableName() string { return "demo_receipts" }
