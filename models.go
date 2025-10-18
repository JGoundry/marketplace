package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

var ErrInsufficientFunds error = errors.New("insufficient funds")
var ErrNoURL error = errors.New("need to set PG_URL env var")

const TokenLength = 32

type UserPurchase struct {
	username    string
	itemName    string
	itemPrice   int
	purchasedAt time.Time
}

func (u UserPurchase) String() string {
	return fmt.Sprintf("username: %v, item: %v, price: %v, time: %v", u.username, u.itemName, convertMoneyPrintable(u.itemPrice), u.purchasedAt.String())
}

type User struct {
	userId       int
	username     string
	passwordHash string
	balance      int
	lastLogin    time.Time
	createdAt    time.Time
}

type Session struct {
	sessionId  string
	csrfToken  string
	userId     int
	ipAddr     []byte
	expires_at time.Time
}

type Item struct {
	itemId      int
	name        string
	description string
	price       int
}

func (i Item) String() string {
	return fmt.Sprintf("id: %v, name: %v, description: %v, price: %v", i.itemId, i.name, i.description, convertMoneyPrintable(i.price))
}

type Purchase struct {
	purchaseId  int
	userId      int
	itemId      int
	purchasedAt time.Time
}

type DB interface {
	Items() ([]Item, error)
	Purchases(userId int) ([]UserPurchase, error)
	GetUserFromUsername(username string) (User, error)
	GetItem(itemId int) (Item, error)
	Register(username, passwordHash string) (User, error)
	CreateSession(user User, ipAddr string) (Session, error)
	GetSession(sessionId string) (Session, error)
	UpdateLastLogin(userId int)
	Balance(userId int) (int, error)
	Deposit(userId int, amount int) (int, error)
	Purchase(userId int, itemId int) error
	Close() error
}

type SqlDB struct {
	db   *sql.DB
	wg   sync.WaitGroup
	done chan struct{}
}

func NewSqlDB() (*SqlDB, error) {
	dsn := os.Getenv("PG_URL")
	if len(dsn) == 0 {
		return nil, ErrNoURL
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	} else if err := db.Ping(); err != nil {
		return nil, err
	}

	sqlDb := SqlDB{
		db:   db,
		done: make(chan struct{}),
	}

	// Database cleanup operation on seperate goroutine - possibly move this up to env for logging purposes
	sqlDb.wg.Add(1)
	go func() {
		for {
			select {
			case <-time.After(24 * time.Hour):
				sqlDb.RemoveExpiredSessions()
			case <-sqlDb.done:
				return
			}
		}
	}()

	return &sqlDb, nil
}

func (s *SqlDB) Close() error {
	s.done <- struct{}{}
	s.wg.Wait()
	return nil
}

func (s *SqlDB) Items() ([]Item, error) {
	query := `SELECT item_id, name, description, CAST(price*100 AS INT) FROM items`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	var items []Item
	var item Item

	for rows.Next() {
		err := rows.Scan(&item.itemId, &item.name, &item.description, &item.price)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *SqlDB) Purchases(userId int) ([]UserPurchase, error) {
	query := `SELECT users.username, items.name, CAST(purchases.price*100 AS INT), purchases.purchased_at
			  FROM users
			  JOIN purchases ON users.user_id=purchases.user_id
			  JOIN items ON purchases.item_id=items.item_id
			  WHERE users.user_id=$1`

	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}

	var purchases []UserPurchase
	var purchase UserPurchase // declare here so we dont allocate each time

	for rows.Next() {
		err := rows.Scan(&purchase.username, &purchase.itemName, &purchase.itemPrice, &purchase.purchasedAt)
		if err != nil {
			return nil, err
		}
		purchases = append(purchases, purchase)
	}

	return purchases, nil
}

func scanUser(row *sql.Row) (User, error) {
	var user User
	err := row.Scan(&user.userId, &user.username, &user.passwordHash, &user.balance, &user.lastLogin, &user.createdAt)
	return user, err
}

func scanItem(row *sql.Row) (Item, error) {
	var item Item
	err := row.Scan(&item.itemId, &item.name, &item.description, &item.price)
	return item, err
}

func scanSession(row *sql.Row) (Session, error) {
	var session Session
	err := row.Scan(&session.sessionId, &session.csrfToken, &session.userId, &session.ipAddr, &session.expires_at)
	return session, err
}

func (s *SqlDB) CreateSession(user User, ipAddr string) (Session, error) {
	var err error
	var sessionId string
	var csrfToken string

	// Generate unique session id
	for {
		sessionId, err = generateToken(TokenLength)
		if err != nil {
			return Session{}, err
		}
		if _, err = s.GetSession(sessionId); err != nil {
			break // break if unique
		}
	}

	csrfToken, err = generateToken(TokenLength)
	if err != nil {
		return Session{}, nil
	}

	query := `INSERT INTO sessions (session_id, csrf_token, user_id, ip_addr, expires_at) 
			  VALUES ($1, $2, $3, $4, $5)
			  RETURNING session_id, csrf_token, user_id, ip_addr, expires_at`

	row := s.db.QueryRow(query, sessionId, csrfToken, user.userId, ipAddr, time.Now().Add(time.Hour*24))
	return scanSession(row)
}

func (s *SqlDB) GetUserFromUsername(username string) (User, error) {
	query := `SELECT user_id, username, password_hash, CAST(balance*100 AS INT), last_login, created_at FROM users WHERE username=$1`
	row := s.db.QueryRow(query, username)
	return scanUser(row)
}

func (s *SqlDB) GetItem(itemId int) (Item, error) {
	query := `SELECT item_id, name, description, CAST(price*100 AS INT) FROM items WHERE item_id=$1`
	row := s.db.QueryRow(query, itemId)
	return scanItem(row)
}

func (s *SqlDB) GetSession(sessionId string) (Session, error) {
	query := `SELECT * FROM sessions WHERE session_id=$1`
	row := s.db.QueryRow(query, sessionId)
	return scanSession(row)
}

func (s *SqlDB) Register(username, passwordHash string) (User, error) {
	var user User
	var err error
	query := `INSERT INTO users (username, password_hash)
	 		  VALUES ($1, $2) 
			  RETURNING user_id, username, password_hash, CAST(balance*100 AS INT), last_login, created_at`
	row := s.db.QueryRow(query, username, passwordHash)
	err = row.Scan(&user.userId, &user.username, &user.passwordHash, &user.balance, &user.lastLogin, &user.createdAt)
	return user, err
}

func (s *SqlDB) UpdateLastLogin(userId int) {
	query := `UPDATE users SET last_login=NOW() WHERE user_id=$1`
	s.db.Exec(query, userId)
}

func (s *SqlDB) RemoveExpiredSessions() (sql.Result, error) {
	query := `DELETE FROM sessions WHERE expires_at<NOW()`
	return s.db.Exec(query)
}

func (s *SqlDB) Balance(userId int) (int, error) {
	var balance int
	query := `SELECT CAST(users.balance*100 AS INT) FROM users WHERE user_id=$1`
	row := s.db.QueryRow(query, userId)
	err := row.Scan(&balance)
	return balance, err
}

func (s *SqlDB) Deposit(userId int, amount int) (int, error) {
	var balance int
	query := `UPDATE users SET balance=balance+CAST($1 AS NUMERIC(10, 2))/100 WHERE users.user_id=$2 RETURNING CAST(users.balance*100 AS INT)`
	row := s.db.QueryRow(query, amount, userId)
	err := row.Scan(&balance)
	return balance, err
}

func (s *SqlDB) Purchase(userId int, itemId int) (err error) {
	var tx *sql.Tx
	tx, err = s.db.Begin()
	if err != nil {
		return err
	}

	// Rollback or commit depending on err
	defer func() {
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get item price
	var price int
	getPriceQuery := `SELECT CAST(price*100 AS INT) FROM items WHERE items.item_id=$1 FOR UPDATE`
	row := tx.QueryRow(getPriceQuery, itemId)
	err = row.Scan(&price)
	if err != nil {
		return err
	}

	// Get user balance
	var balance int
	getBalanceQuery := `SELECT CAST(balance*100 AS INT) FROM users WHERE users.user_id=$1 FOR UPDATE`
	row = tx.QueryRow(getBalanceQuery, userId)
	err = row.Scan(&balance)
	if err != nil {
		return err
	}

	// Check for sufficient funds
	if balance < price {
		return ErrInsufficientFunds
	}

	// Subtract price from balance
	updateBalanceQuery := `UPDATE users SET balance=balance-CAST($1 AS NUMERIC(10,2))/100 WHERE users.user_id=$2`
	_, err = tx.Exec(updateBalanceQuery, price, userId)
	if err != nil {
		return err
	}

	// Create purchase
	addPurchaseQuery := `INSERT INTO purchases (user_id, item_id, price) VALUES ($1, $2, CAST($3 AS NUMERIC(10, 2))/100)`
	_, err = tx.Exec(addPurchaseQuery, userId, itemId, price)
	return err
}
