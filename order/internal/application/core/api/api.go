package api

import (
	"github.com/lailaaquino/microservices/order/internal/application/core/domain"
	"github.com/lailaaquino/microservices/order/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db      ports.DBPort
	payment ports.PaymentPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort) *Application {
	return &Application{
		db:      db,
		payment: payment,
	}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
	var totalItems int32 = 0
	for _, item := range order.OrderItems {
		totalItems += item.Quantity
	}

	if totalItems > 50 {
		return domain.Order{}, status.Errorf(codes.InvalidArgument, "Order failed: Total quantity of items (%d) exceeds the maximum limit of 50.", totalItems)
	}
	err := a.db.Save(&order)
	if err != nil {
		return domain.Order{}, err
	}

	paymentErr := a.payment.Charge(&order)
	if paymentErr != nil {
		order.Status = "Canceled"
		_ = a.db.Save(&order)
		return domain.Order{}, paymentErr
	}
	order.Status = "Paid"
	err = a.db.Save(&order)
	if err != nil {
		return domain.Order{}, err
	}

	return order, nil
}
