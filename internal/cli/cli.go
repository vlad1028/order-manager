package cli

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/vlad1028/order-manager/internal/models/order"
	"io"
	"strconv"
	"strings"
	"sync"
)

type OrderManagerCLI struct {
	in         *bufio.Reader
	out        io.Writer
	outMu      sync.Mutex
	rootCmd    *cobra.Command
	adaptor    OrderCLIAdaptor
	workerPool *WorkerPool
}

type OrderCLIAdaptor interface {
	AcceptOrder(req *AcceptOrderRequest) error
	CancelOrder(req *CancelOrderRequest) error
	IssueOrder(req *IssueOrderRequest) ([]*order.Order, error)
	GetOrders(req *GetOrdersRequest) ([]*order.Order, error)
	AcceptReturn(req *AcceptReturnRequest) error
	GetReturned(req *GetReturnedRequest) ([]*order.Order, error)
}

func NewOrderManagerCLI(a OrderCLIAdaptor, r io.Reader, w io.Writer) *OrderManagerCLI {
	c := &cobra.Command{
		Use:   "",
		Short: "Order Manager for Pickup Point",
	}

	sh := &OrderManagerCLI{
		in:         bufio.NewReader(r),
		out:        w,
		rootCmd:    c,
		adaptor:    a,
		workerPool: NewWorkerPool(2, 5),
	}
	sh.addCommands()

	return sh
}

func (r *OrderManagerCLI) addCommands() {
	r.rootCmd.AddCommand(
		r.newAcceptOrderCmd(),
		r.newCancelOrderCmd(),
		r.newIssueOrderCmd(),
		r.newGetOrdersCmd(),
		r.newAcceptReturnCmd(),
		r.newGetReturnedCmd(),
		r.newSetWorkersCmd(),
	)
}

func (r *OrderManagerCLI) readLine() string {
	input, _ := r.in.ReadString('\n')
	return strings.TrimSpace(input)
}

func (r *OrderManagerCLI) writeOut(s string) {
	r.outMu.Lock()
	_, _ = r.out.Write([]byte(s))
	r.outMu.Unlock()
}

func (r *OrderManagerCLI) writeErr(err error) {
	r.writeOut(fmt.Sprintf("Error: %s\n> ", err)) // out and err are the same
}

func (r *OrderManagerCLI) printfln(format string, a ...interface{}) {
	var msg = fmt.Sprintf(format, a...)
	r.writeOut(msg + "\n> ")
}

func (r *OrderManagerCLI) println(a ...interface{}) {
	var msg = fmt.Sprintln(a...)
	r.writeOut(msg + "> ")
}

func (r *OrderManagerCLI) print(a ...interface{}) {
	var msg = fmt.Sprint(a...)
	r.writeOut(msg)
}

func (r *OrderManagerCLI) RunInteractive() {
	r.println("Order Manager Interactive Shell")
	r.print("Type 'help' to see available commands or 'exit' to quit.\n")

	for {
		print("> ")
		input := r.readLine()

		if input == "exit" {
			r.println("Exiting...")
			r.Shutdown()
			return
		}

		err := r.Run(input)
		if err != nil {
			r.writeErr(err)
		}
	}
}

func (r *OrderManagerCLI) Shutdown() {
	r.println("Shutting down gracefully...")
	r.workerPool.Close()
	r.println("All tasks completed. Exiting.")
}

func (r *OrderManagerCLI) Run(input string) error {
	args := strings.Split(input, " ")
	r.rootCmd.SetArgs(args)
	err := r.rootCmd.Execute()
	resetFlagsRecursively(r.rootCmd)
	return err
}

func resetFlagsRecursively(c *cobra.Command) {
	c.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = flag.Value.Set(flag.DefValue)
	})
	for _, cmd := range c.Commands() {
		resetFlagsRecursively(cmd)
	}
}

func (r *OrderManagerCLI) newSetWorkersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-workers [numWorkers]",
		Short: "Set the number of workers in the pool",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			numWorkers := args[0]

			n, err := strconv.Atoi(numWorkers)
			if err != nil {
				r.writeErr(err)
			} else {
				if n < 0 {
					r.writeErr(fmt.Errorf("number of workers must be non-negative"))
				}
				r.workerPool.SetNumWorkers(uint(n))
				r.printfln("Number of workers set to %d", n)
			}
		},
	}
}

func (r *OrderManagerCLI) newAcceptOrderCmd() *cobra.Command {
	var pack string
	var addFilm bool

	cmd := &cobra.Command{
		Use:   "accept-order [orderID] [clientID] [weight] [cost]",
		Short: "Accept an order delivery from adaptor courier",
		Args:  cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			req := &AcceptOrderRequest{
				ID:        args[0],
				ClientID:  args[1],
				Weight:    args[2],
				Cost:      args[3],
				Packaging: pack,
				AddFilm:   addFilm,
			}

			go func() {
				if err := r.adaptor.AcceptOrder(req); err != nil {
					r.writeErr(err)
				} else {
					r.println("Order accepted.")
				}
			}()
		},
	}

	cmd.Flags().StringVarP(&pack, "package", "p", "", "Order package")
	cmd.Flags().BoolVarP(&addFilm, "firm", "f", false, "Add additional firm")

	return cmd
}

func (r *OrderManagerCLI) newCancelOrderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel-order [orderID]",
		Short: "Return an order to the courier and cancel",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req := &CancelOrderRequest{
				ID: args[0],
			}

			r.workerPool.AddTask(func() {
				if err := r.adaptor.CancelOrder(req); err != nil {
					r.writeErr(err)
				} else {
					r.println("Order cancelled.")
				}
			})
		},
	}
}

func (r *OrderManagerCLI) newIssueOrderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "issue-order [orderIDs...]",
		Short: "Issue orders to adaptor client",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req := &IssueOrderRequest{
				IDs: args,
			}

			r.workerPool.AddTask(func() {
				orders, err := r.adaptor.IssueOrder(req)
				if err != nil {
					r.writeErr(err)
				}
				r.printfln("Orders issued: %v", orders)
			})
		},
	}
}

func (r *OrderManagerCLI) newGetOrdersCmd() *cobra.Command {
	var limit int
	var localOnly bool

	cmd := &cobra.Command{
		Use:   "get-orders [clientID]",
		Short: "Get orders for adaptor client",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req := &GetOrdersRequest{
				ClientID:  args[0],
				LocalOnly: localOnly,
			}

			r.workerPool.AddTask(func() {
				if orders, err := r.adaptor.GetOrders(req); err != nil {
					r.writeErr(err)
					return
				} else {
					r.paginateOrders(orders, limit)
				}
			})
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", -1, "Limit the number of orders returned")
	cmd.Flags().BoolVarP(&localOnly, "local", "l", false, "Show only orders that are in this Pick Up Point")

	return cmd
}

func (r *OrderManagerCLI) paginateOrders(orders []*order.Order, limit int) {
	if limit == -1 {
		limit = len(orders)
	}

	for i := 0; i < len(orders); i += limit {
		end := min(i+limit, len(orders))

		r.printOrders(orders[i:end])

		if end == len(orders) || !r.promptForMore() {
			break
		}
	}
}

func (r *OrderManagerCLI) printOrders(orders []*order.Order) {
	for _, o := range orders {
		r.printfln("Order ID: %d, Delivered At: %s", o.ID, o.StatusUpdated)
	}
}

func (r *OrderManagerCLI) promptForMore() bool {
	r.print("Press any key to see more orders, or type 'q' to quit: ")
	return r.readLine() != "q"
}

func (r *OrderManagerCLI) newAcceptReturnCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "accept-return [clientID] [orderID]",
		Short: "Accept an order return from adaptor client",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			req := &AcceptReturnRequest{
				ClientID: args[0],
				OrderID:  args[1],
			}

			r.workerPool.AddTask(func() {
				if err := r.adaptor.AcceptReturn(req); err != nil {
					r.writeErr(err)
				} else {
					r.println("Order returned.")
				}
			})
		},
	}
}

func (r *OrderManagerCLI) newGetReturnedCmd() *cobra.Command {
	var page int
	var perPage int

	cmd := &cobra.Command{
		Use:   "get-returned",
		Short: "Get returned orders with pagination",
		Run: func(cmd *cobra.Command, args []string) {
			req := &GetReturnedRequest{
				Page:    page,
				PerPage: perPage,
			}

			r.workerPool.AddTask(func() {
				returns, err := r.adaptor.GetReturned(req)
				if err != nil {
					r.writeErr(err)
					return
				}

				for _, ret := range returns {
					r.printfln("Return ID: %d, Return Date: %s", ret.ID, ret.StatusUpdated)
				}
			})
		},
	}

	cmd.Flags().IntVarP(&page, "page", "p", 0, "Page number")
	cmd.Flags().IntVarP(&perPage, "per-page", "n", 10, "Number of returns per page")

	return cmd
}
