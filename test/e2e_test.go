package test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vlad1028/order-manager/internal/cache"
	"github.com/vlad1028/order-manager/internal/db"
	"github.com/vlad1028/order-manager/internal/kafka"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderRepo "github.com/vlad1028/order-manager/internal/order"
	"github.com/vlad1028/order-manager/internal/order/service"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/vlad1028/order-manager/internal/cli"
)

type OrderManagerSuite struct {
	suite.Suite
	input  *bytes.Buffer
	output *bytes.Buffer
	shell  *cli.OrderManagerCLI
	repo   orderRepo.Repository
	db     *pgxpool.Pool
}

func (suite *OrderManagerSuite) SetupSuite() {
	suite.input = new(bytes.Buffer)
	suite.output = new(bytes.Buffer)

	pool, err := db.ConnectToDB("test")
	suite.Require().NoError(err)
	suite.db = pool
	suite.repo = db.SetupOrderRepository(pool)

	orderService := service.NewOrderService(0, 24*7*time.Hour, 2*24*time.Hour, suite.repo, kafka.NewMockProducer(), cache.NewCacheMock())
	orderHandler := cli.NewOrderServiceAdaptor(orderService)

	suite.shell = cli.NewOrderManagerCLI(orderHandler, suite.input, suite.output)
}

func (suite *OrderManagerSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *OrderManagerSuite) SetupTest() {
	_, err := suite.db.Exec(context.Background(), "TRUNCATE TABLE orders RESTART IDENTITY CASCADE")
	suite.Require().NoError(err)
}

func (suite *OrderManagerSuite) putOrder(o *order.Order) {
	input := fmt.Sprintf("accept-order %d %d %d %d", o.ID, o.ClientID, o.Weight, o.Cost)
	suite.send(input)
}

func (suite *OrderManagerSuite) send(in string) {
	fmt.Println("Input: " + in)
	suite.Require().NoError(suite.shell.Run(in))
}

func (suite *OrderManagerSuite) TestAddOrder() {
	o := order.NewOrder(1, 1, 0, 10, 15)

	suite.putOrder(o)

	suite.Contains(suite.output.String(), "Order accepted.")

	gotOrder, err := suite.repo.Get(context.Background(), 1)
	suite.Require().NoError(err)
	gotOrder.StatusUpdated = o.StatusUpdated
	suite.Equal(o, gotOrder)
}

func (suite *OrderManagerSuite) TestGetOrder() {
	o := &order.Order{
		ID:       2,
		ClientID: 2,
		Weight:   10,
		Cost:     15,
	}

	suite.putOrder(o)
	suite.send(fmt.Sprintf("get-orders %d", o.ClientID))

	suite.Contains(suite.output.String(), fmt.Sprintf("Order ID: %d", o.ID))
}

func TestOrderManagerSuite(t *testing.T) {
	suite.Run(t, new(OrderManagerSuite))
}
