package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

const txnCtx = "txn-ctx"

func setTxnContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txnCtx, tx)
}

func getTxnFromContext(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txnCtx).(*sqlx.Tx)
	return tx
}

// Transact wraps sql operations in a transaction.
func Transact(ctx context.Context, db *sqlx.DB, txFunc func(context.Context, *sqlx.Tx) error) error {

	_, err := TransactType(ctx, db, func(txnCtx context.Context, tx *sqlx.Tx) (struct{}, error) {
		return struct{}{}, txFunc(txnCtx, tx)
	})
	return err
}

func TransactType[T any](ctx context.Context, db *sqlx.DB, txFunc func(context.Context, *sqlx.Tx) (T, error)) (T, error) {
	var v T
	tx := getTxnFromContext(ctx)
	isRoot := false
	var err error
	if tx == nil {

		isRoot = true
		tx, err = db.Beginx()
		if err != nil {
			return v, err
		}
		ctx = setTxnContext(ctx, tx)
	}

	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback()
			panic(p)
		}
		if err != nil {
			if inErr := tx.Rollback(); inErr != nil && !errors.Is(inErr, sql.ErrTxDone) {
				err = inErr
			}
		} else if isRoot { // only commit if we are the one who actually injected the context
			txnErr := tx.Commit()
			if err != nil && !errors.Is(txnErr, sql.ErrTxDone) {
				err = txnErr
			}
		}
	}()

	v, err = txFunc(ctx, tx)
	return v, err
}
