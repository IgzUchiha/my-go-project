package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "my-go-project/app"

    "go.temporal.io/sdk/client"
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
    _, err = c.ExecuteWorkflow(context.Background(), options, app.CartWorkflowExample, state) // use = instead of :=
    if err != nil {
        log.Fatalln("unable to execute workflow", err)
    }
}