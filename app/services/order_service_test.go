package services_test

import (
	"context"
	"errors"
	"testing"

	"prismgo-demo/app/demo/testing"
	"prismgo-demo/app/models"
	"prismgo-demo/app/repositories"
	"prismgo-demo/app/services"

	"github.com/prismgo/framework/database"
	"gorm.io/gorm"
)

func TestOrderServiceCreateAndPay(t *testing.T) {
	app := demotest.NewApplication(t, demotest.Options{})
	db := database.Resolve()
	if err := db.AutoMigrate(&models.User{}, &models.Order{}, &models.Receipt{}); err != nil {
		t.Fatalf("migrate demo models: %v", err)
	}
	user := &models.User{Name: "Demo User", Email: "demo@example.test", Password: "not-a-real-hash"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create demo user: %v", err)
	}

	raw, err := app.Container().Make(services.OrderServiceKey)
	if err != nil {
		t.Fatalf("resolve order service: %v", err)
	}
	service := raw.(*services.OrderService)
	order, err := service.Create(context.Background(), user.ID, 2599)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.Status != models.OrderStatusPending || order.ID == 0 {
		t.Fatalf("created order = %#v", order)
	}

	paid, err := service.Pay(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("pay order: %v", err)
	}
	if paid.Status != models.OrderStatusPaid || paid.PaidAt == nil {
		t.Fatalf("paid order = %#v", paid)
	}
	if _, err := service.Pay(context.Background(), order.ID); !errors.Is(err, repositories.ErrOrderAlreadyPaid) {
		t.Fatalf("second payment error = %v", err)
	}
}

func TestOrderServiceValidatesTotalAndMissingOrder(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	db := database.Resolve()
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		t.Fatal(err)
	}
	service := services.NewOrderService(repositories.NewOrderRepository(db))
	if _, err := service.Create(context.Background(), 1, 0); !errors.Is(err, services.ErrInvalidOrderTotal) {
		t.Fatalf("invalid total error = %v", err)
	}
	if _, err := service.Find(context.Background(), 404); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing order error = %v", err)
	}
}
