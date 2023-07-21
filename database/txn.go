package database

import "github.com/jmoiron/sqlx"

// Transact wraps sql operations in a transaction.
func Transact(db *sqlx.DB, txFunc func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback()
			panic(p)
		}
		if err != nil {
			if inErr := tx.Rollback(); inErr != nil {
				err = inErr
			}
		} else {
			err = tx.Commit()
		}
	}()
	err = txFunc(tx)
	return err
}
