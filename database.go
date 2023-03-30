package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Database struct {
	connect *sql.DB
	config  Config
}

type Buy struct {
	Id           int64
	Symbol       string
	Coins        float64
	ExchangeRate float64
	DesiredPrice float64
	CreatedAt    string
	RealOrderId  int64
	RealQuantity float64
	HasSellOrder int64
}

func NewDatabase(config Config) Database {
	//name := time.Now().Format("db/db_2006_01_02__15_04_05.db")
	name := ":memory:"

	if IS_REAL_ENABLED {
		name = time.Now().Format("db/real_2006_01_02__15_04_05.db")

		if USE_REAL_MONEY {
			name = REAL_MONEY_DB_NAME + ".db"
		}
	}
	connect, _ := sql.Open("sqlite3", name)

	createBuysTable(connect)
	createSellsTable(connect)

	return Database{
		connect: connect,
		config:  config,
	}
}

func (db *Database) Close() {
	db.connect.Close()
}

func createBuysTable(connect *sql.DB) sql.Result {
	query := `
		CREATE TABLE IF NOT EXISTS buys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol VARCHAR(255),
			coins FLOAT,
			exchange_rate FLOAT,
		    desired_price FLOAT,
			created_at DATETIME,
		    real_order_id INTEGER,
			real_quantity FLOAT,
		    has_sell_order INTEGER
		);
	`
	result, err := connect.Exec(query)
	if err != nil {
		panic(err)
	}

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
	result, err := connect.Exec(query)
	if err != nil {
		panic(err)
	}

	return result
}

// User functions

func (db *Database) AddBuy(symbol string, coinsCount, exchangeRate, desiredPrice float64, createdAt string) sql.Result {
	query := `
		INSERT INTO buys (symbol, coins, exchange_rate, desired_price, created_at, real_order_id, real_quantity, has_sell_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	result, err := db.connect.Exec(query, symbol, coinsCount, exchangeRate, desiredPrice, createdAt, 0, 0.0, 0)
	if err != nil {
		panic(err)
	}

	return result
}

func (db *Database) AddRealBuy(symbol string, coinsCount, exchangeRate, desiredPrice float64, createdAt string, orderId int64, quantity float64) sql.Result {
	//createdAt := time.Now().Format("2006-01-02 15:04:05")
	query := `
		INSERT INTO buys (symbol, coins, exchange_rate, desired_price, created_at, real_order_id, real_quantity, has_sell_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`

	result, err := db.connect.Exec(query, symbol, coinsCount, exchangeRate, desiredPrice, createdAt, orderId, quantity, 0)
	if err != nil {
		panic(err)
	}

	return result
}

func (db *Database) UpdateRealBuyOrderId(buyId int64, orderId int64) {
	query := `
		UPDATE buys
		SET real_order_id = $1, has_sell_order = 1
		WHERE id = $2
	`

	_, err := db.connect.Exec(query, orderId, buyId)
	if err != nil {
		panic(err)
	}
}

func (db *Database) AddSell(
	symbol string,
	coinsCount float64,
	exchangeRate float64,
	revenue float64,
	buyId int64,
	createdAt string,
) sql.Result {
	//createdAt := time.Now().Format("2006-01-02 15:04:05")
	query := `
		INSERT INTO sells (symbol, coins, exchange_rate, revenue, buy_id, created_at) VALUES ($1, $2, $3, $4, $5, $6);
	`
	result, err := db.connect.Exec(query, symbol, coinsCount, exchangeRate, revenue, buyId, createdAt)
	if err != nil {
		panic(err)
	}

	return result
}

func (db *Database) FetchUnsoldBuysByUpperPercentage(exchangeRate, upperPercentage float64) []Buy {
	unsoldBuys := []Buy{}
	query := `
		SELECT b.*
		FROM buys AS b 
        LEFT JOIN sells AS s 
        	ON s.buy_id = b.id 
        WHERE s.id IS NULL 
            AND (b.exchange_rate + ((b.exchange_rate * $1) / 100)) <= $2   
	`

	rows, err := db.connect.Query(query, upperPercentage, exchangeRate)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		buy := Buy{}
		rows.Scan(&buy.Id, &buy.Symbol, &buy.Coins, &buy.ExchangeRate, &buy.DesiredPrice, &buy.CreatedAt, &buy.RealOrderId, &buy.RealQuantity, &buy.HasSellOrder)
		unsoldBuys = append(unsoldBuys, buy)
	}

	return unsoldBuys
}

func (db *Database) FetchUnsoldBuysByLowerPercentage(exchangeRate, lowerPercentage float64) []Buy {
	unsoldBuys := []Buy{}
	query := `
		SELECT b.*
		FROM buys AS b 
        LEFT JOIN sells AS s 
        	ON s.buy_id = b.id 
        WHERE s.id IS NULL 
            AND (b.exchange_rate - ((b.exchange_rate * $1) / 100)) >= $2   
	`

	rows, err := db.connect.Query(query, lowerPercentage, exchangeRate)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		buy := Buy{}
		rows.Scan(&buy.Id, &buy.Symbol, &buy.Coins, &buy.ExchangeRate, &buy.DesiredPrice, &buy.CreatedAt, &buy.RealOrderId, &buy.RealQuantity, &buy.HasSellOrder)
		unsoldBuys = append(unsoldBuys, buy)
	}

	return unsoldBuys
}

func (db *Database) FetchUnsoldBuysByDesiredPrice(exchangeRate float64) []Buy {
	unsoldBuys := []Buy{}
	query := `
		SELECT b.*
		FROM buys AS b 
        LEFT JOIN sells AS s 
        	ON s.buy_id = b.id 
        WHERE s.id IS NULL 
            AND b.desired_price <= $1  
	`

	rows, err := db.connect.Query(query, exchangeRate)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		buy := Buy{}
		rows.Scan(&buy.Id, &buy.Symbol, &buy.Coins, &buy.ExchangeRate, &buy.DesiredPrice, &buy.CreatedAt, &buy.RealOrderId, &buy.RealQuantity, &buy.HasSellOrder)
		unsoldBuys = append(unsoldBuys, buy)
	}

	return unsoldBuys
}

func (db *Database) FetchUnsoldBuys() []Buy {
	unsoldBuys := []Buy{}
	query := `
		SELECT b.*
		FROM buys AS b 
        LEFT JOIN sells AS s 
        	ON s.buy_id = b.id 
        WHERE s.id IS NULL
	`

	rows, err := db.connect.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		buy := Buy{}
		rows.Scan(&buy.Id, &buy.Symbol, &buy.Coins, &buy.ExchangeRate, &buy.DesiredPrice, &buy.CreatedAt, &buy.RealOrderId, &buy.RealQuantity, &buy.HasSellOrder)
		unsoldBuys = append(unsoldBuys, buy)
	}

	return unsoldBuys
}

func (db *Database) FetchUnsoldBuysById(buyIds []int64) []Buy {
	unsoldBuys := []Buy{}
	query := fmt.Sprintf(`
		SELECT b.*
		FROM buys AS b 
        LEFT JOIN sells AS s 
        	ON s.buy_id = b.id 
        WHERE s.id IS NULL AND b.id IN(%s)
	`, JoinInt64(buyIds))

	rows, err := db.connect.Query(query, JoinInt64(buyIds))
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		buy := Buy{}
		rows.Scan(&buy.Id, &buy.Symbol, &buy.Coins, &buy.ExchangeRate, &buy.DesiredPrice, &buy.CreatedAt, &buy.RealOrderId, &buy.RealQuantity, &buy.HasSellOrder)
		unsoldBuys = append(unsoldBuys, buy)
	}

	return unsoldBuys
}

type revenue struct {
	value float64
}

func (db *Database) GetTotalRevenue() float64 {
	rev := revenue{}
	query := `
		SELECT (SUM(revenue) - COUNT(id) * $1) AS rev 
		FROM sells 
		GROUP BY symbol
	`
	row := (*db).connect.QueryRow(query, db.config.TotalMoneyAmount*LEVERAGE)
	row.Scan(&rev.value)

	return rev.value
}

type buysCount struct {
	value int
}

func (db *Database) GetBuysCount() int {
	count := buysCount{}
	query := `
		SELECT COUNT(id) AS c 
		FROM buys 
	`
	row := (*db).connect.QueryRow(query)
	row.Scan(&count.value)

	return count.value
}

func (db *Database) CountUnsoldBuys() int {
	var count int
	query := `
		SELECT COUNT(b.id)
		FROM buys AS b 
        LEFT JOIN sells AS s 
        	ON s.buy_id = b.id 
        WHERE s.id IS NULL
	`
	(*db).connect.QueryRow(query).Scan(&count)

	return count
}

func (db *Database) GetAvgSellTime() float64 {
	var sellTime float64
	query := `
		SELECT AVG(JULIANDAY(s.created_at) - JULIANDAY(b.created_at))
		FROM buys AS b
        INNER JOIN sells AS s ON b.id = s.buy_id
	`
	row := db.connect.QueryRow(query)
	row.Scan(&sellTime)

	if row.Err() != nil {
		panic(row.Err())
	}

	return sellTime
}

func (db *Database) GetMedianSellTime() float64 {
	var sellTimes []float64
	query := `
		SELECT (JULIANDAY(s.created_at) - JULIANDAY(b.created_at))
		FROM buys AS b
        INNER JOIN sells AS s ON b.id = s.buy_id
	`
	rows, err := db.connect.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var sellTime float64
		rows.Scan(&sellTime)
		sellTimes = append(sellTimes, sellTime)
	}

	return Median(sellTimes)
}

func (db *Database) CanBuyInGivenPeriod(createdAt string, period int) bool {
	var count int
	query := `
		SELECT COUNT(id)
		FROM buys
        WHERE created_at > $1
	`

	candleTime := ConvertDateStringToTime(createdAt)
	canNotBuyDuration := GetCurrentMinusTime(candleTime, period)
	db.connect.QueryRow(query, canNotBuyDuration).Scan(&count)

	return count == 0
}
