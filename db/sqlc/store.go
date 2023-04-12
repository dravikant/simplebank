package db

import (
	"context"
	"database/sql"
	"fmt"
)

// store provides all functionality to execute database queries and Transactions
// this is a composition -- all functions provided by Queries are available so that trasactions can be implemented
type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx Error %v,rb Error %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()

}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	Transfer    Transfer `json:"transfer"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// perform money trnasfer from one account to other
// add entry to transfer add entry to both account update balance to both accounts
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {

	var result TransferTxResult

	// Convert the TransferTxParams object to a CreateTransferParams object
	params := CreateTransferParams(arg)

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		//create a transfer entry

		result.Transfer, err = q.CreateTransfer(ctx, params)

		fmt.Println(">> Inside TransferTx :accounts from and to:", result.FromAccount.ID, result.ToAccount.ID)
		if err != nil {
			return err
		}

		//create an entry for from_account
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		//create an entry for from_account
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		//TODO: update the balance in both accounts

		return nil
	})

	return result, err
}
