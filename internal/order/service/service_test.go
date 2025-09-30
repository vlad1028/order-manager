package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/vlad1028/order-manager/internal/cache"
	"github.com/vlad1028/order-manager/internal/kafka"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderInterfaces "github.com/vlad1028/order-manager/internal/order"
	"github.com/vlad1028/order-manager/internal/order/repository/mock"
	"testing"
	"time"
)

func newTestService(r orderInterfaces.Repository) *Service {
	return newTestServiceWithMessageSender(r, kafka.NewMockProducer())
}

func newTestServiceWithMessageSender(r orderInterfaces.Repository, producer MessageSender) *Service {
	return NewOrderService(0, 24*7*time.Hour, 2*24*time.Hour, r, producer, cache.NewCacheMock())
}

func TestOrderService_GetOrders(t *testing.T) {
	type mockResults struct {
		get []*order.Order
	}
	request := orderInterfaces.GetOrdersRequest{
		ClientID:  1,
		LocalOnly: false,
	}
	exampleOrders := []*order.Order{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	ctrl := minimock.NewController(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		request     *orderInterfaces.GetOrdersRequest
		mockResults mockResults
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			"Success",
			&request,
			mockResults{
				get: exampleOrders,
			},
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mock.NewOrderRepositoryMock(ctrl)
			orderRepo.GetByMock.Return(tt.mockResults.get, nil)

			m := newTestService(orderRepo)
			_, err := m.GetOrders(ctx, &request)
			tt.wantErr(t, err)
		})
	}
}

func TestOrderService_GetReturned(t *testing.T) {
	type mockResults struct {
		get []*order.Order
	}

	ctrl := minimock.NewController(t)
	ctx := context.Background()

	request := orderInterfaces.GetReturnedRequest{Page: 0, PerPage: -1}
	exampleOrders := []*order.Order{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	tests := []struct {
		name        string
		request     *orderInterfaces.GetReturnedRequest
		mockResults mockResults
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			"Success",
			&request,
			mockResults{
				get: exampleOrders,
			},
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mock.NewOrderRepositoryMock(ctrl)
			orderRepo.GetByPaginatedMock.Return(tt.mockResults.get, nil)

			m := newTestService(orderRepo)
			_, gotErr := m.GetReturned(ctx, tt.request)
			tt.wantErr(t, gotErr)
		})
	}
}

func TestOrderService_IssueOrder(t *testing.T) {
	type mockResults struct {
		get    *order.Order
		getby  []*order.Order
		add    error
		remove error
	}

	ctrl := minimock.NewController(t)
	ctx := context.Background()

	request := orderInterfaces.IssueOrderRequest{
		IDs: []basetypes.ID{1, 2, 3},
	}

	exampleOrdersSuccess := []*order.Order{
		{ID: 1, Status: order.Stored},
		{ID: 2, Status: order.Stored},
		{ID: 3, Status: order.Stored},
	}
	exampleOrdersError := []*order.Order{
		{ID: 4, Status: order.Stored},
		{ID: 5, Status: order.Stored},
		{ID: 6, Status: order.Stored},
	}

	tests := []struct {
		name        string
		request     *orderInterfaces.IssueOrderRequest
		mockResults mockResults
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			"Success",
			&request,
			mockResults{
				get:    exampleOrdersSuccess[0],
				getby:  exampleOrdersSuccess,
				add:    nil,
				remove: nil,
			},
			assert.NoError,
		},
		{
			"RepoError",
			&request,
			mockResults{
				get:    exampleOrdersError[0],
				getby:  exampleOrdersError,
				add:    fmt.Errorf("error"),
				remove: nil,
			},
			assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mock.NewOrderRepositoryMock(ctrl)
			orderRepo.GetMock.Return(tt.mockResults.get, nil)
			orderRepo.GetByMock.Return(tt.mockResults.getby, nil)
			orderRepo.AddOrUpdateListMock.Return(tt.mockResults.add)

			m := newTestService(orderRepo)
			_, gotErr := m.IssueOrder(ctx, tt.request)
			tt.wantErr(t, gotErr)
		})
	}
}

func TestOrderService_AcceptOrder(t *testing.T) {
	type mockResult struct {
		exists bool
		err    error
	}

	ctrl := minimock.NewController(t)
	ctx := context.Background()

	exampleOrderRequest := orderInterfaces.AcceptOrderRequest{
		ID:        1,
		ClientID:  1,
		Weight:    10,
		Cost:      10,
		Packaging: nil,
		AddFilm:   false,
	}

	tests := []struct {
		name       string
		request    *orderInterfaces.AcceptOrderRequest
		mockResult mockResult
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			"Success",
			&exampleOrderRequest,
			mockResult{
				false,
				nil,
			},
			assert.NoError,
		},
		{
			"Exists",
			&exampleOrderRequest,
			mockResult{
				true,
				nil,
			},
			assert.Error,
		},
		{
			"ReposError",
			&exampleOrderRequest,
			mockResult{
				false,
				fmt.Errorf("error"),
			},
			assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mock.NewOrderRepositoryMock(ctrl)
			orderRepo.AddOrUpdateMock.Optional().Return(tt.mockResult.exists, tt.mockResult.err)

			m := newTestService(orderRepo)
			_, err := m.AcceptOrder(ctx, tt.request)
			tt.wantErr(t, err)
		})
	}
}

func TestOrderService_AcceptReturn(t *testing.T) {
	type mockResults struct {
		get    *order.Order
		add    error
		remove error
	}

	ctrl := minimock.NewController(t)
	ctx := context.Background()

	returnRequest := orderInterfaces.AcceptReturnRequest{
		ClientID: 1,
		OrderID:  1,
	}
	exampleOrder := order.Order{
		ID:            1,
		ClientID:      1,
		PickupPointID: 0,
		Status:        order.ReachedClient,
		StatusUpdated: time.Now(),
	}
	var expiredRefundOrder = exampleOrder
	expiredRefundOrder.StatusUpdated = time.Now().AddDate(0, 0, -10)
	var wrongClientID = exampleOrder
	wrongClientID.ClientID = 100000
	var wrongPickUpPoint = exampleOrder
	wrongPickUpPoint.PickupPointID = 100000

	tests := []struct {
		name        string
		request     *orderInterfaces.AcceptReturnRequest
		mockResults mockResults
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			"Success",
			&returnRequest,
			mockResults{
				get:    &exampleOrder,
				add:    nil,
				remove: nil,
			},
			assert.NoError,
		},
		{
			"TooLateToMakeRefund",
			&returnRequest,
			mockResults{
				get:    &expiredRefundOrder,
				add:    nil,
				remove: nil,
			},
			assert.Error,
		},
		{
			"WrongClientID",
			&returnRequest,
			mockResults{
				get:    &wrongClientID,
				add:    nil,
				remove: nil,
			},
			assert.Error,
		},
		{
			"WrongPickupPointID",
			&returnRequest,
			mockResults{
				get:    &wrongPickUpPoint,
				add:    nil,
				remove: nil,
			},
			assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mock.NewOrderRepositoryMock(ctrl)
			orderRepo.GetMock.Optional().Return(tt.mockResults.get, nil)
			orderRepo.AddOrUpdateMock.Optional().Return(false, tt.mockResults.add)
			orderRepo.DeleteMock.Optional().Return(tt.mockResults.remove)

			m := newTestService(orderRepo)
			_, err := m.AcceptReturn(ctx, tt.request)
			tt.wantErr(t, err)
		})
	}
}

func TestOrderService_CancelOrder(t *testing.T) {
	type mockResults struct {
		get    *order.Order
		add    error
		remove error
	}

	ctrl := minimock.NewController(t)
	ctx := context.Background()

	cancelOrderRequest := orderInterfaces.CancelOrderRequest{
		ID: 1,
	}
	exampleOrder := order.Order{
		ID:            1,
		Status:        order.Returned,
		StatusUpdated: time.Now(),
	}
	var cannotBeCanceled = exampleOrder
	cannotBeCanceled.Status = order.Stored
	cannotBeCanceled.StatusUpdated = time.Now() // not expired

	tests := []struct {
		name        string
		request     *orderInterfaces.CancelOrderRequest
		mockResults mockResults
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			"Success",
			&cancelOrderRequest,
			mockResults{
				get:    &exampleOrder,
				add:    nil,
				remove: nil,
			},
			assert.NoError,
		},
		{
			"CannotBeReturned",
			&cancelOrderRequest,
			mockResults{
				get:    &cannotBeCanceled,
				add:    nil,
				remove: nil,
			},
			assert.Error,
		},
		{
			"RepoError",
			&cancelOrderRequest,
			mockResults{
				get:    &exampleOrder,
				add:    fmt.Errorf("error"),
				remove: nil,
			},
			assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mock.NewOrderRepositoryMock(ctrl)
			orderRepo.GetMock.Optional().Return(tt.mockResults.get, nil)
			orderRepo.AddOrUpdateMock.Optional().Return(false, tt.mockResults.add)
			orderRepo.DeleteMock.Optional().Return(tt.mockResults.remove)

			m := newTestService(orderRepo)
			_, err := m.CancelOrder(ctx, tt.request)
			tt.wantErr(t, err)
		})
	}
}

func TestSendEvent(t *testing.T) {
	ctrl := minimock.NewController(t)
	mockProducer := kafka.NewMockProducer()
	mockRepo := mock.NewOrderRepositoryMock(ctrl)
	orderService := newTestServiceWithMessageSender(mockRepo, mockProducer)

	events := []order.Event{
		{
			OrderID:   basetypes.ID(123),
			Operation: "accept",
			Timestamp: time.Now().UTC(),
		},
		{
			OrderID:   basetypes.ID(124),
			Operation: "return",
			Timestamp: time.Now().UTC(),
		},
		{
			OrderID:   basetypes.ID(125),
			Operation: "issue",
			Timestamp: time.Now().UTC(),
		},
	}

	for _, e := range events {
		orderService.sendEvent(e)
	}

	assert.Equal(t, len(events), len(mockProducer.Messages))

	for i, e := range events {
		msg := mockProducer.Messages[i]
		var receivedEvent order.Event
		err := json.Unmarshal(msg.Value, &receivedEvent)
		assert.NoError(t, err)
		assert.Equal(t, e.OrderID, receivedEvent.OrderID)
		assert.Equal(t, e.Operation, receivedEvent.Operation)
		assert.Equal(t, e.Timestamp, receivedEvent.Timestamp)
	}
}
