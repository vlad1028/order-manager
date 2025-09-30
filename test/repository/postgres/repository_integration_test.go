package postgres

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/vlad1028/order-manager/internal/db"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderRepo "github.com/vlad1028/order-manager/internal/order"
	"math/rand"
	"testing"
)

type OrderRepositoryTestSuite struct {
	suite.Suite
	ctx  context.Context
	db   *pgxpool.Pool
	repo orderRepo.Repository
}

func (suite *OrderRepositoryTestSuite) SetupSuite() {
	pool, err := db.ConnectToDB("test")
	suite.Require().NoError(err)
	suite.db = pool
	repo := db.SetupOrderRepository(pool)
	suite.repo = repo
}

func (suite *OrderRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *OrderRepositoryTestSuite) SetupTest() {
	suite.ctx = context.Background()
	_, err := suite.db.Exec(suite.ctx, "TRUNCATE TABLE orders RESTART IDENTITY CASCADE")
	suite.Require().NoError(err)
}

func TestOrderRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrderRepositoryTestSuite))
}

func randomOrderStatus() order.Status {
	statuses := []order.Status{order.Returned, order.Stored, order.ReachedClient, order.Canceled}
	return statuses[rand.Intn(len(statuses))]
}

func generateFakeOrder() *order.Order {
	//gofakeit.Seed(time.Now().UnixNano())
	return &order.Order{
		ID:       basetypes.ID(gofakeit.Int64()),
		ClientID: basetypes.ID(gofakeit.Int64()),
		Status:   randomOrderStatus(),
		Weight:   uint(gofakeit.IntN(1000)),
		Cost:     uint(gofakeit.IntN(1000)),
	}
}

func (suite *OrderRepositoryTestSuite) TestAddOrUpdate() {
	fakeOrder := generateFakeOrder()
	ctx := context.Background()

	exists, err := suite.repo.AddOrUpdate(ctx, fakeOrder)
	suite.Require().NoError(err)
	suite.Require().False(exists, "Order should not exist initially")

	exists, err = suite.repo.AddOrUpdate(ctx, fakeOrder)
	suite.Require().NoError(err)
	suite.Require().True(exists, "Order should exist after adding")
}

func (suite *OrderRepositoryTestSuite) TestGet() {
	fakeOrder := generateFakeOrder()
	ctx := context.Background()

	_, err := suite.repo.AddOrUpdate(ctx, fakeOrder)
	suite.Require().NoError(err)

	fetchedOrder, err := suite.repo.Get(ctx, fakeOrder.ID)
	suite.Require().NoError(err)
	suite.Require().NotNil(fetchedOrder)
	suite.Require().Equal(fakeOrder.ID, fetchedOrder.ID)
}

func (suite *OrderRepositoryTestSuite) TestDelete() {
	fakeOrder := generateFakeOrder()
	ctx := context.Background()

	_, err := suite.repo.AddOrUpdate(ctx, fakeOrder)
	suite.Require().NoError(err)

	err = suite.repo.Delete(ctx, fakeOrder.ID)
	suite.Require().NoError(err)

	deletedOrder, err := suite.repo.Get(ctx, fakeOrder.ID)
	suite.Require().Error(err)
	suite.Require().Nil(deletedOrder)
}

func (suite *OrderRepositoryTestSuite) TestAddOrUpdateList() {
	orders := []*order.Order{
		generateFakeOrder(),
		generateFakeOrder(),
		generateFakeOrder(),
	}
	ctx := context.Background()

	err := suite.repo.AddOrUpdateList(ctx, orders)
	suite.Require().NoError(err)

	for _, o := range orders {
		fetchedOrder, err := suite.repo.Get(ctx, o.ID)
		suite.Require().NoError(err)
		suite.Require().NotNil(fetchedOrder)
		suite.Require().Equal(o.ID, fetchedOrder.ID)
	}
}

func (suite *OrderRepositoryTestSuite) TestGetBy() {
	orders := []*order.Order{
		generateFakeOrder(),
		generateFakeOrder(),
		generateFakeOrder(),
	}
	ctx := context.Background()

	err := suite.repo.AddOrUpdateList(ctx, orders)
	suite.Require().NoError(err)

	filter := &order.Filter{ClientID: &orders[0].ClientID}
	fetchedOrders, err := suite.repo.GetBy(ctx, filter)
	suite.Require().NoError(err)
	suite.Require().Len(fetchedOrders, 1)
	suite.Require().Equal(orders[0].ClientID, fetchedOrders[0].ClientID)
}
