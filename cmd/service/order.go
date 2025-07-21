package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/khatibomar/durable_execution_toy/internal/engine"
)

type OrderService struct{}

type Order struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
}

func processPayment(order Order) (string, error) {
	time.Sleep(100 * time.Millisecond)
	paymentRef := fmt.Sprintf("PAY_%s_%d", order.ID, int(order.Amount*100))
	fmt.Printf("Payment processed: %s\n", paymentRef)
	return paymentRef, nil
}

func reserveInventory(order Order) error {
	if rand.Float32() < 0.5 {
		return errors.New("inventory system down for maintenance")
	}

	time.Sleep(80 * time.Millisecond)
	fmt.Printf("Reserved %d units of %s\n", order.Quantity, order.Product)
	return nil
}

func sendConfirmation(order Order, paymentRef string) error {
	if rand.Float32() < 0.5 {
		return errors.New("email service temporarily down")
	}

	time.Sleep(50 * time.Millisecond)
	fmt.Printf("Confirmation about payment %s sent to user %s\n", paymentRef, order.UserID)
	return nil
}

func (s OrderService) ProcessOrder(ctx engine.Context, order Order) error {
	fmt.Printf("\nProcessing order %s\n", order.ID)

	paymentRef, err := engine.Run(ctx, func() (string, error) {
		return processPayment(order)
	})
	if err != nil {
		return fmt.Errorf("payment failed: %w", err)
	}

	_, err = engine.Run(ctx, func() (engine.Void, error) {
		return engine.Void{}, reserveInventory(order)
	})
	if err != nil {
		return fmt.Errorf("inventory reservation failed: %w", err)
	}

	_, err = engine.Run(ctx, func() (engine.Void, error) {
		return engine.Void{}, sendConfirmation(order, paymentRef)
	})
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}

	fmt.Printf("Order %s completed successfully\n", order.ID)
	return nil
}
