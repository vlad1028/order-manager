package postgres

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
	errors "github.com/vlad1028/order-manager/internal/order"
	"reflect"
	"strings"
)

type PgRepository struct {
}

func NewPgRepository() *PgRepository {
	return &PgRepository{}
}

func (r *PgRepository) Get(ctx context.Context, tx pgx.Tx, id basetypes.ID) (*order.Order, error) {
	var o order.Order
	err := pgxscan.Get(ctx, tx, &o,
		"SELECT (id, client_id, pickup_point_id, status, status_updated, weight, cost) FROM orders WHERE id = $1",
		id)

	if err != nil {
		return nil, err
	}

	return &o, err
}

func (r *PgRepository) Delete(ctx context.Context, tx pgx.Tx, id basetypes.ID) error {
	result, err := tx.Exec(ctx,
		"DELETE FROM orders WHERE id = $1",
		id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.ErrOrderNotFound
	}

	return nil
}

func (r *PgRepository) AddOrUpdate(ctx context.Context, tx pgx.Tx, o *order.Order) (exists bool, err error) {
	query := `
        INSERT INTO orders (id, client_id, pickup_point_id, status, weight, cost, status_updated)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (id)
		DO UPDATE SET 
			status = excluded.cost,
			status_updated = excluded.status_updated
    RETURNING (xmax != 0) AS exists;
    `

	err = tx.QueryRow(
		ctx, query,
		o.ID, o.ClientID, o.PickupPointID, o.Status, o.Weight, o.Cost,
	).Scan(&exists)

	return exists, err
}

func (r *PgRepository) GetBy(ctx context.Context, tx pgx.Tx, filter *order.Filter) ([]*order.Order, error) {
	return r.GetByPaginated(ctx, tx, filter, 0, -1)
}

func (r *PgRepository) GetByPaginated(ctx context.Context, tx pgx.Tx, filter *order.Filter, offset uint, limit int) ([]*order.Order, error) {
	var orders []*order.Order
	query, args := buildFilterQueryPaginated(filter, "SELECT (id, client_id, pickup_point_id, status, status_updated, weight, cost)", offset, limit)
	err := pgxscan.Select(ctx, tx, &orders, query, args...)

	return orders, err
}

func (r *PgRepository) DeleteBy(ctx context.Context, tx pgx.Tx, filter *order.Filter) error {
	query, args := buildFilterQuery(filter, "DELETE")
	_, err := tx.Exec(ctx, query, args...)

	return err
}

func buildFilterQuery(filter *order.Filter, operation string) (string, []interface{}) {
	return buildFilterQueryPaginated(filter, operation, 0, -1)
}

func buildFilterQueryPaginated(filter *order.Filter, operation string, offset uint, limit int) (string, []interface{}) {
	query := operation + " FROM orders WHERE"
	var conditions []string
	var args []interface{}

	v := reflect.ValueOf(filter).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)
		columnName := fieldType.Tag.Get("db")

		if !fieldValue.IsNil() {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", columnName, len(args)+1))
			args = append(args, fieldValue.Interface())
		}
	}

	if len(conditions) > 0 {
		query += " " + strings.Join(conditions, " AND ")
	} else {
		query += " TRUE" // no filters
	}

	if limit > 0 {
		query += fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}

	return query, args
}
