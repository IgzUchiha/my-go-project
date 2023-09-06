package abandoned_cart

import (
    // "context"
    // "fmt"
    // "github.com/mailgun/mailgun-go"
    // "os"
	"time"
	"go.temporal.io/sdk/workflow"
)

// ... (other imports and variable declarations)
const abandonedCartTimeout = 10 * time.Minute // Adjust the duration as needed


// CartWorkflow is your workflow function
func CartWorkflow(ctx workflow.Context, state CartState) error {
    logger := workflow.GetLogger(ctx)

    err := workflow.SetQueryHandler(ctx, "getCart", func(input []byte) (CartState, error) {
        return state, nil
    })
    if err != nil {
        logger.Info("SetQueryHandler failed.", "Error", err)
        return err
    }

    channel := workflow.GetSignalChannel(ctx, "cartMessages")
    sentAbandonedCartEmail := false

    for {
        // Create a new Selector on each iteration of the loop means Temporal will pick the first
        // event that occurs each time: either receiving a signal, or responding to the timer.
        selector := workflow.NewSelector(ctx)
        selector.AddReceive(channel, func(c workflow.ReceiveChannel, _ bool) {
            var signal interface{}
            c.Receive(ctx, &signal)

            // Handle signals for updating the cart
        })

        // If the user doesn't update the cart for `abandonedCartTimeout`, send an email
        // reminding them about their cart. Only send the email once.
        if !sentAbandonedCartEmail && len(state.Items) > 0 {
            selector.AddFuture(workflow.NewTimer(ctx, abandonedCartTimeout), func(f workflow.Future) {
                sentAbandonedCartEmail = true
                ao := workflow.ActivityOptions{
                    StartToCloseTimeout: 10 * time.Second,
                }

                ctx = workflow.WithActivityOptions(ctx, ao)

                // More on SendAbandonedCartEmail in the next section
                err := workflow.ExecuteActivity(ctx, SendAbandonedCartEmail, state.Email).Get(ctx, nil)
                if err != nil {
                    logger.Error("Error sending email %v", err)
                    return
                }
            })
        }

        selector.Select(ctx)
    }

    return nil
}
