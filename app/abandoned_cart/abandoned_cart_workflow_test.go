package abandoned_cart

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)



type AbandonedCartUnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func (s *AbandonedCartUnitTestSuite) SetupTest() {
	// Initialize Temporal's testing utilities and setup here
	s.WorkflowTestSuite.SetupTest()
}

func (s *AbandonedCartUnitTestSuite) AfterTest(suiteName, testName string) {
	// Clean up and assert expectations here
	s.WorkflowTestSuite.AfterTest(suiteName, testName)
}

func TestAbandonedCartUnitTestSuite(t *testing.T) {
	suite.Run(t, new(AbandonedCartUnitTestSuite))
}

func (s *AbandonedCartUnitTestSuite) Test_AbandonedCart() {
	cart := CartState{Items: make([]CartItem, 0)}

	// Mock the SendAbandonedCartEmail activity
	var a *Activities
	sendTo := ""
	s.env.OnActivity(a.SendAbandonedCartEmail, mock.Anything, mock.Anything).Return(
		func(_ context.Context, _sendTo string) error {
			sendTo = _sendTo
			return nil
		})

	// Add a product to the cart and update the email
	s.env.RegisterDelayedCallback(func() {
		update := AddToCartSignal{
			Route: RouteTypes.ADD_TO_CART,
			Item:  CartItem{ProductId: 1, Quantity: 1},
		}
		s.env.SignalWorkflow("cartMessages", update)

		updateEmail := UpdateEmailSignal{
			Route: RouteTypes.UPDATE_EMAIL,
			Email: "abandoned_test@temporal.io",
		}
		s.env.SignalWorkflow("cartMessages", updateEmail)
	}, time.Millisecond*1)

	// Wait for 10 mins and make sure abandoned cart email has been sent. The extra
	// 2ms is because signals are async, so the last change to the cart happens at 2ms.
	s.env.RegisterDelayedCallback(func() {
		s.Equal(sendTo, "abandoned_test@temporal.io")
	}, abandonedCartTimeout+time.Millisecond*2)

	s.env.ExecuteWorkflow(AbandonedCartWorkflow, cart)

	s.True(s.env.IsWorkflowCompleted())
}
