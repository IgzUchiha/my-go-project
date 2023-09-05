package app

import (
    "go.temporal.io/sdk/workflow"
    // "my-go-project/app"
)

type (
    CartItem struct {
        ProductId int
        Quantity  int
    }

    CartState struct {
        Items []CartItem
        Email string
    }
)

func CartWorkflowExample(ctx workflow.Context, state CartState) error {
    logger := workflow.GetLogger(ctx)

    err := workflow.SetQueryHandler(ctx, "getCart", func(input []byte) (CartState, error) {
        return state, nil
    })
    if err != nil {
        logger.Info("SetQueryHandler failed.", "Error", err)
        return err
    }

    channel := workflow.GetSignalChannel(ctx, "cartMessages")
    selector := workflow.NewSelector(ctx)

    selector.AddReceive(channel, func(c workflow.ReceiveChannel, _ bool) {
        var signal interface{}
        c.Receive(ctx, &signal)
        state.Items = append(state.Items, CartItem{ProductId: 0, Quantity: 1})
    })

    for {
        selector.Select(ctx)
    }

    return nil
}