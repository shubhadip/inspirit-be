package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Models struct {
	DB DBModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Password     string    `json:"-"`
	WalletAmount float64   `json:"wallet_amount"`
	BitcoinValue float64   `json:"bitcoin_value"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type Wallet struct {
	gorm.Model
	ID        int `json:"id"`
	UserId    int `json:"user_id"`
	User      User
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Bitcoin struct {
	gorm.Model
	ID           int `json:"id"`
	UserId       int `json:"user_id"`
	WalletId     int
	Wallet       Wallet
	User         User
	Type         string    `json:"type"`
	Amount       float64   `json:"amount"`
	CurrentPrice float64   `json:"current_price"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func InitialMigration(DB *gorm.DB) {
	DB.AutoMigrate(&User{})
	DB.AutoMigrate(&Wallet{})
	DB.AutoMigrate(&Bitcoin{})
}
