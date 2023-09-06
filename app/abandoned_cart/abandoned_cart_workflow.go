package abandoned_cart

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
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
			timeoutDuration := 10 * time.Minute // Send email after 10 minutes of inactivity
			selector.AddFuture(workflow.NewTimer(ctx, timeoutDuration), func(f workflow.Future) {
				sentAbandonedCartEmail = true

				// Define and initialize the ActivityOptions (ao)
				ao := workflow.ActivityOptions{
					StartToCloseTimeout: 10 * time.Second,
				}

				// Send the abandoned cart email (replace with your email sending code)
				ctx = workflow.WithActivityOptions(ctx, ao)
				err := workflow.ExecuteActivity(ctx, SendAbandonedCartEmail, state.Email).Get(ctx, nil)
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
	// Create an instance of the Mailgun client
	domain := "your-domain.com"
	apiKey := "your-api-key"
	mg := mailgun.NewMailgun(domain, apiKey)

	// Create a new email message
	message := mg.NewMessage(
		"no-reply@"+domain, // From
		"Your cart is waiting for you!", // Subject
		"Hi there, we noticed that you've left some items in your cart. Why not come back and complete your purchase?", // Body
		email, // To
	)

	// Send the email
	_, _, err := mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Printf("Sent abandoned cart email to %s\n", email)
	return nil
}

func main() {
	c, err := client.NewClient(client.Options{})
	if err != nil {
		fmt.Println("Error creating Temporal client:", err)
		return
	}
	defer c.Close()

	// Register the activity with the Temporal client
	c.RegisterActivity(SendAbandonedCartEmail)

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
	if err != nil