package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Database struct {
	connect *sql.DB
}

type Buy struct {
	Id           int
	Symbol       string
	Coins        float64
	ExchangeRate float64
	CreatedAt    string
}

func NewDatabase() Database {
	name := time.Now().Format("db/db_2006_01_02__15_04_05.db")
	//name := ":memory:"
	connect, _ := sql.Open("sqlite3", name)

	createBuysTable(connect)
	createSellsTable(connect)

	return Database{
		connect: connect,
	}
}

func (db *Database) Close() {
	db.connect.Close()
}

func createBuysTable(connect *sql.DB) sql.Result {
	query := `
		CREATE TABLE IF NOT EXISTS buys (
			id integer primary key AUTOINCREMENT,
			symbol VARCHAR(255),
			coins FLOAT,
			exchange_rate FLOAT,
			created_at DATETIME
		);
	`
	result, _ := connect.Exec(query)

	return result
}

func createSellsTable(connect *sql.DB) sql.Result {
	query := `
		CREATE TABLE IF NOT EXISTS sells (
			id integer primary key AUTOINCREMENT,
			symbol VARCHAR(255),
			coins FLOAT,
			exchange_rate FLOAT,
			revenue FLOAT,
			buy_id INT,
			created_at DATETIME
		);
	`
	result, _ := connect.Exec(query)

	return result
}

// User functions

func (db *Database) AddBuy(symbol string, coinsCount float64, exchangeRate float64, createdAt string) sql.Result {
	query := `
		INSERT INTO buys (symbol, coins, exchange_rate, created_at) VALUES ($1, $2, $3, $4);
	`
	result, _ := db.connect.Exec(query, symbol, coinsCount, exchangeRate, createdAt)
	return result
}

func (db *Database) AddSell(
	symbol string,
	coinsCount float64,
	exchangeRate float64,
	revenue float64,
	buyId int,
	createdAt string,
) sql.Result {
	//createdAt := time.Now().Format("2006-01-02 15:04:05")
	query := `
		INSERT INTO sells (symbol, coins, exchange_rate, revenue, buy_id, created_at) VALUES ($1, $2, $3, $4, $5, $6);
	`
	result, _ := db.connect.Exec(query, symbol, exchangeRate, coinsCount, revenue, buyId, createdAt)
	return result
}
