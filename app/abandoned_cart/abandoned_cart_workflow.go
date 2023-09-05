// main.go

package main

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type CartItem struct {
	ProductID int
	Quantity  int
}

type CartState struct {
	Items []CartItem
	Email string
}

func main() {
	c, err := client.NewClient(client.Options{})
	if err != nil {
		fmt.Println("Error creating Temporal client:", err)
		return
	}
	defer c.Close()

	// Start the abandoned cart workflow
	options := client.StartWorkflowOptions{
		ID:        "ABANDONED_CART-1", // Unique workflow ID
		TaskQueue: "ABANDONED_CART_TASK_QUEUE",
	}

	state := CartState{Email: "user@example.com"} // Set the user's email
	we, err := c.ExecuteWorkflow(context.Background(), options, AbandonedCartWorkflow, state)
	if err != nil {
		fmt.Println("Error starting workflow:", err)
		return
	}

	// Wait for the workflow to complete (in this example, it runs indefinitely)
	var result interface{}
	err = we.Get(context.Background(), &result)
	if err != nil {
		fmt.Println("Error getting workflow result:", err)
		return
	}
}

func AbandonedCartWorkflow(ctx workflow.Context, state CartState) error {
	logger := workflow.GetLogger(ctx)

	// Set up a query handler to retrieve the cart state
	err := workflow.SetQueryHandler(ctx, "getCart", func(input []byte) (CartState, error) {
		return state, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed.", "Error", err)
		return err
	}

	// Create a new Selector to handle signals and timers
	channel := workflow.GetSignalChannel(ctx, "cartMessages")
	sentAbandonedCartEmail := false

	for {
		selector := workflow.NewSelector(ctx)

		// Listen for signals (cart updates)
		selector.AddReceive(channel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)

			// Handle signals for updating the cart (e.g., append items)
			// Replace this logic with your cart update code
			state.Items = append(state.Items, CartItem{ProductID: 0, Quantity: 1})
		})

		// If the user doesn't update the cart for a specified duration, send an email reminder
		if !sentAbandonedCartEmail && len(state.Items) > 0 {
			timeoutDuration := 3 * time.Hour // Adjust the duration as needed
			selector.AddFuture(workflow.NewTimer(ctx, timeoutDuration), func(f workflow.Future) {
				sentAbandonedCartEmail = true

				// Send the abandoned cart email (replace with your email sending code)
				err := SendAbandonedCartEmail(ctx, state.Email)
				if err != nil {
					logger.Error("Error sending email:", err)
				}
			})
		}

		// Wait for either a signal or a timer to elapse
		selector.Select(ctx)
	}

	return nil
}

func SendAbandonedCartEmail(ctx context.Context, email string) error {
	// Implement your email sending logic here
	// Replace this placeholder code with your actual email sending code
	fmt.Printf("Sending abandoned cart email to %s\n", email)
	return nil
}
