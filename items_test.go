package main

// import (
// 	"log"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// type TestDB struct{}

// func (t TestDB) Items() ([]Item, error) {
// 	return []Item{Item{}}, nil
// }

// func NewTestEnv() Env {
// 	return Env{
// 		logger: log.Default(),
// 		db:     TestDB{},
// 	}
// }

// func TestItems(t *testing.T) {
// 	r := httptest.NewRequest("GET", "/items", nil)
// 	w := httptest.NewRecorder()

// 	e := NewTestEnv()
// 	e.Items(w, r)

// 	result := w.Result()
// 	defer result.Body.Close()

// 	if result.StatusCode != http.StatusOK {
// 		t.FailNow()
// 	}
// }
