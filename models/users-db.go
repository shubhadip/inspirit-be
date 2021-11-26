package models

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)

type DBModel struct {
	DB *sql.DB
}

func (m *DBModel) Get(id int) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, username, wallet_amount, bitcoin_value from users where id =$1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var user User

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.WalletAmount,
		&user.BitcoinValue,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (m *DBModel) InsertUser(user User) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stmt := `insert into users (username, password, wallet_amount, bitcoin_value,
				created_at, updated_at) values (?,?,?,?,?,?)`
	log.Println(stmt)

	_, err := m.DB.Exec(stmt,
		user.Username,
		user.Password,
		user.WalletAmount,
		user.BitcoinValue,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		log.Println("err")
		return err
	}

	return nil
}

func (m *DBModel) GetUser(username string) (*User, error) {
	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `select id, username, password from users where username = ?`
	row := m.DB.QueryRow(query, username)
	var user User

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &user, nil

}

func (m *DBModel) GetUserById(id int) (*User, error) {
	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	log.Println(id)
	query := `select id, username, wallet_amount, bitcoin_value from users where id = ?`
	row := m.DB.QueryRow(query, id)
	var user User

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.WalletAmount,
		&user.BitcoinValue,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &user, nil

}

func (m *DBModel) PurchaseBitCoin(id int64, amount float64, currentBitCoinValue float64) (*User, error) {
	var bitCoinValue float64 = currentBitCoinValue
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, username, wallet_amount, bitcoin_value from users where id = ?`
	row := m.DB.QueryRow(query, id)
	var user User

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.WalletAmount,
		&user.BitcoinValue,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if user.WalletAmount < amount {
		return nil, errors.New("invalid amount entered")
	}

	var wallet Wallet
	var bitcoin Bitcoin

	wallet.UserId = user.ID
	wallet.Type = "debit"
	wallet.Amount = amount
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	tx, err := m.DB.BeginTx(ctx, nil)

	result, err := tx.ExecContext(ctx, `INSERT INTO wallets (user_id, type, amount, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		wallet.UserId, wallet.Type, wallet.Amount, wallet.CreatedAt, wallet.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	walletId, err := result.LastInsertId()

	bitcoin.UserId = user.ID
	bitcoin.WalletId = int(walletId)
	bitcoin.Type = "purchase"
	bitcoin.Amount = amount
	bitcoin.CurrentPrice = bitCoinValue
	bitcoin.CreatedAt = time.Now()
	bitcoin.UpdatedAt = time.Now()

	_, errB := tx.ExecContext(ctx, `INSERT INTO bitcoins (user_id, wallet_id, type, amount, current_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		bitcoin.UserId, int(walletId), bitcoin.Type, bitcoin.Amount, bitcoin.CurrentPrice, bitcoin.CreatedAt, bitcoin.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return nil, errB
	}

	tempBitCoinValue := amount / bitCoinValue

	_, err = tx.ExecContext(ctx, "UPDATE users SET bitcoin_value = bitcoin_value + ? WHERE id = ?",
		tempBitCoinValue, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	result, err = tx.ExecContext(ctx, "UPDATE users SET wallet_amount = wallet_amount - ? WHERE id = ?",
		amount, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	updatedRow := m.DB.QueryRow(query, id)

	errValue := updatedRow.Scan(
		&user.ID,
		&user.Username,
		&user.WalletAmount,
		&user.BitcoinValue,
	)

	if errValue != nil {
		log.Println(errValue)
		return nil, errValue
	}

	return &user, nil
}

func (m *DBModel) SellBitCoin(id int64, value float64, currentBitCoinValue float64) (*User, error) {
	var bitCoinValue float64 = currentBitCoinValue
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, username, wallet_amount, bitcoin_value from users where id = ?`
	row := m.DB.QueryRow(query, id)
	var user User

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.WalletAmount,
		&user.BitcoinValue,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if user.BitcoinValue < value {
		return nil, errors.New("invalid amount entered")
	}

	var actualBitcoinAmount = value * bitCoinValue

	var wallet Wallet
	var bitcoin Bitcoin

	wallet.UserId = user.ID
	wallet.Type = "credit"
	wallet.Amount = actualBitcoinAmount
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	tx, err := m.DB.BeginTx(ctx, nil)

	result, err := tx.ExecContext(ctx, `INSERT INTO wallets (user_id, type, amount, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		wallet.UserId, wallet.Type, wallet.Amount, wallet.CreatedAt, wallet.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	walletId, err := result.LastInsertId()

	bitcoin.UserId = user.ID
	bitcoin.WalletId = int(walletId)
	bitcoin.Type = "sell"
	bitcoin.Amount = actualBitcoinAmount
	bitcoin.CurrentPrice = bitCoinValue
	bitcoin.CreatedAt = time.Now()
	bitcoin.UpdatedAt = time.Now()

	_, errB := tx.ExecContext(ctx, `INSERT INTO bitcoins (user_id, wallet_id, type, amount, current_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		bitcoin.UserId, int(walletId), bitcoin.Type, bitcoin.Amount, bitcoin.CurrentPrice, bitcoin.CreatedAt, bitcoin.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return nil, errB
	}

	_, err = tx.ExecContext(ctx, "UPDATE users SET bitcoin_value = bitcoin_value - ? WHERE id = ?",
		value, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "UPDATE users SET wallet_amount = wallet_amount + ? WHERE id = ?",
		actualBitcoinAmount, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	updatedRow := m.DB.QueryRow(query, id)

	errValue := updatedRow.Scan(
		&user.ID,
		&user.Username,
		&user.WalletAmount,
		&user.BitcoinValue,
	)

	if errValue != nil {
		log.Println(errValue)
		return nil, errValue
	}

	return &user, nil
}
