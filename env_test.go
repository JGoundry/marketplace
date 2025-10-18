package main

// import (
// 	"log"
// 	"time"
// )

// // import (
// // 	"log"
// // 	"net/http"
// // 	"net/http/httptest"
// // 	"testing"
// // )

// type TestDB struct {
// 	users     []User
// 	items     []Item
// 	purchases []Purchase
// 	sessions  []Session
// }

// var users []User = []User{
// 	User{
// 		userId:       1,
// 		username:     "test_user",
// 		passwordHash: "password",
// 		balance:      0.0,
// 		lastLogin:    time.Now(),
// 		createdAt:    time.Now(),
// 	},
// 	User{
// 		userId:       1,
// 		username:     "rich_test_user",
// 		passwordHash: "password",
// 		balance:      17000.0,
// 		lastLogin:    time.Now(),
// 		createdAt:    time.Now(),
// 	},
// }

// var items []Item = []Item{
// 	Item{
// 		itemId:      1,
// 		name:        "Nvidia RTX 3060",
// 		description: "Graphics card",
// 		price:       170.0,
// 	},
// }

// func (t TestDB) Items() ([]Item, error)                                  {}
// func (t TestDB) Purchases(userId int) ([]UserPurchase, error)            {}
// func (t TestDB) GetUserFromUsername(username string) (User, error)       {}
// func (t TestDB) GetUserFromId(userId int) (User, error)                  {}
// func (t TestDB) GetItem(itemId int) (Item, error)                        {}
// func (t TestDB) Register(username, passwordHash string) (User, error)    {}
// func (t TestDB) CreateSession(user User, ipAddr string) (Session, error) {}
// func (t TestDB) GetSession(sessionId string) (Session, error)            {}
// func (t TestDB) UpdateLastLogin(userId int)                              {}
// func (t TestDB) Balance(userId int) (float64, error)                     {}
// func (t TestDB) Deposit(userId int, amount float64) (float64, error)     {}
// func (t TestDB) Purchase(userId int, itemId int) error                   {}
// func (t TestDB) Close() error                                            {}

// func NewTestEnv() Env {
// 	return Env{
// 		logger: log.Default(),
// 		db:     TestDB{},
// 	}
// }

// // func TestItems(t *testing.T) {
// // 	r := httptest.NewRequest("GET", "/items", nil)
// // 	w := httptest.NewRecorder()

// // 	e := NewTestEnv()
// // 	e.Items(w, r)

// // 	result := w.Result()
// // 	defer result.Body.Close()

// // 	if result.StatusCode != http.StatusOK {
// // 		t.FailNow()
// // 	}
// // }
