package main

import (
    "context"
    "fmt"
    "log"
    "my-go-project/app"
    "my-go-project/app/abandoned_cart" // Import the abandoned cart package
    "time"

    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

func main() {
    c, err := client.NewClient(client.Options{})
    if err != nil {
        log.Fatalln("unable to create Temporal client", err)
    }
    defer c.Close()

    workflowID := "CART-" + fmt.Sprintf("%d", time.Now().Unix())

    options := client.StartWorkflowOptions{
        ID:        workflowID,
        TaskQueue: "CART_TASK_QUEUE",
    }

    state := app.CartState{Items: make([]app.CartItem, 0)}
    _, err = c.ExecuteWorkflow(context.Background(), options, app.CartWorkflowExample, state)
    if err != nil {
        log.Fatalln("unable to execute workflow", err)
    }

    // Create a worker that hosts both Worker and Activity functions
    w := worker.New(c, "CART_TASK_QUEUE", worker.Options{})

    // Register your abandoned cart workflow
    w.RegisterWorkflow(abandoned_cart.AbandonedCartWorkflow)

    // Register the SendAbandonedCartEmail activity
    a := &abandoned_cart.Activities{
        // Configure Mailgun credentials here
    }
    w.RegisterActivity(a.SendAbandonedCartEmailActivity)

    // Start listening to the Task Queue
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Worker execution failed", err)
    }
}
