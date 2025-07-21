package main

import (
	"fmt"

	"github.com/khatibomar/durable_execution_toy/cmd/service"
	"github.com/khatibomar/durable_execution_toy/internal/engine"
)

func main() {
	var executionCtx engine.Context
	var orderService service.OrderService

	for {
		fmt.Println("\nDurable Execution Demo:")
		fmt.Println("1. Process Order")
		fmt.Println("2. Show State")
		fmt.Println("3. Exit")
		fmt.Print("Choice: ")

		var choice int
		fmt.Scan(&choice)

		switch choice {
		case 1:
			if executionCtx == nil {
				executionCtx = engine.NewContext()
			}

			order := service.Order{
				ID:       "ORD-001",
				UserID:   "user-456",
				Amount:   99.99,
				Product:  "Wireless Headphones",
				Quantity: 1,
			}

			err := orderService.ProcessOrder(executionCtx, order)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				fmt.Println("Retry to continue execution.")
			} else {
				fmt.Println("Order processing completed successfully!")
			}
		case 2:
			if executionCtx == nil {
				fmt.Println("No execution found. Process order first.")
				continue
			}
			if ec, ok := executionCtx.(*engine.ExecutionContext); ok {
				ec.PrintState()
			}
		case 3:
			return
		default:
			fmt.Println("Invalid choice.")
		}
	}
}
