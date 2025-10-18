package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type TestDB struct {
	users     []User
	items     []Item
	purchases []Purchase
	sessions  []Session
}

func hashPasswordNoErr(password string) string {
	passwordHash, _ := hashPassword(password)
	return passwordHash
}

var users = []User{
	{
		userId:       1,
		username:     "test_user",
		passwordHash: hashPasswordNoErr("password"),
		balance:      0.0,
		lastLogin:    time.Now(),
		createdAt:    time.Now(),
	},
	{
		userId:       2,
		username:     "rich_test_user",
		passwordHash: hashPasswordNoErr("password"),
		balance:      20000,
		lastLogin:    time.Now(),
		createdAt:    time.Now(),
	},
}

var items = []Item{
	{
		itemId:      1,
		name:        "Nvidia RTX 3060 12GB",
		description: "Graphics Card",
		price:       17500,
	},
}

var sessions = []Session{
	{
		sessionId:  "session",
		csrfToken:  "csrf",
		userId:     1,
		ipAddr:     nil,
		expires_at: time.Now().Add(time.Hour),
	},
	{
		sessionId:  "session",
		csrfToken:  "csrf",
		userId:     1,
		ipAddr:     nil,
		expires_at: time.Now(),
	},
}

var purchases []Purchase

func (t TestDB) Items() ([]Item, error) {
	return items, nil
}

func (t TestDB) Purchases(userId int) ([]UserPurchase, error) {
	var user *User
	var userPurchases []UserPurchase

	// Find user
	for _, currUser := range users {
		if currUser.userId == userId {
			user = &currUser
			break
		}
	}

	// Early return if user not found
	if user == nil {
		return userPurchases, errors.New("user not found")
	}

	// Get purchases
	for _, purchase := range purchases {
		if purchase.userId != userId {
			continue
		}

		// Find item
		var item *Item
		for _, currItem := range items {
			if purchase.itemId == currItem.itemId {
				item = &currItem
				break
			}
		}
		if item == nil {
			continue
		}

		userPurchases = append(userPurchases, UserPurchase{
			user.username,
			item.name,
			item.price,
			purchase.purchasedAt,
		})
	}

	return userPurchases, nil
}

func (t TestDB) GetUserFromUsername(username string) (User, error) {
	for _, user := range users {
		if user.username == username {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}

func (t TestDB) GetItem(itemId int) (Item, error) {
	for _, item := range items {
		if item.itemId == itemId {
			return item, nil
		}
	}
	return Item{}, errors.New("item not found")
}

func (t TestDB) Register(username, passwordHash string) (User, error) {
	// Check name isnt duplicate n get max ID
	id := 0
	for _, user := range users {
		if username == user.username {
			return User{}, errors.New("username already exists")
		}
		id = max(id, user.userId)
	}

	user := User{
		userId:       id + 1,
		username:     username,
		passwordHash: passwordHash,
		balance:      0,
		lastLogin:    time.Now(),
		createdAt:    time.Now(),
	}

	users = append(users, user)
	return user, nil
}

func sessionExists(sessionId string) bool {
	for _, session := range sessions {
		if session.sessionId == sessionId {
			return true
		}
	}
	return false
}

func generateUniqueSessionId() string {
	var sessionId string
	for {
		sessionId, _ = generateToken(TokenLength)
		if !sessionExists(sessionId) {
			break
		}
	}
	return sessionId
}

func (t TestDB) CreateSession(user User, ipAddr string) (Session, error) {
	sessionToken := generateUniqueSessionId()
	csrfToken, _ := generateToken(TokenLength)
	session := Session{
		sessionId:  sessionToken,
		csrfToken:  csrfToken,
		userId:     user.userId,
		ipAddr:     nil,
		expires_at: time.Now().Add(time.Hour * 24),
	}
	sessions = append(sessions, session)
	return session, nil
}
func (t TestDB) GetSession(sessionId string) (Session, error) {
	for _, session := range sessions {
		if session.sessionId == sessionId {
			return session, nil
		}
	}
	return Session{}, errors.New("could not find session")
}
func (t TestDB) Balance(userId int) (int, error) {
	for _, user := range users {
		if user.userId == userId {
			return user.balance, nil
		}
	}
	return 0, errors.New("could not find user")
}
func (t TestDB) Deposit(userId int, amount int) (int, error) {
	for i, user := range users {
		if user.userId == userId {
			user.balance += amount
			users[i] = user
			return user.balance, nil
		}
	}
	return 0, errors.New("could not find user")
}
func (t TestDB) Purchase(userId int, itemId int) error {
	var user *User
	var userIdx int
	for i, currUser := range users {
		if currUser.userId == userId {
			user = &currUser
			userIdx = i
			break
		}
	}
	if user == nil {
		return errors.New("could not find user")
	}

	var item *Item
	for _, currItem := range items {
		if currItem.itemId == itemId {
			item = &currItem
			break
		}
	}
	if item == nil {
		return errors.New("could not find item")
	}

	if user.balance < item.price {
		return ErrInsufficientFunds
	}

	// update user balance
	user.balance -= item.price
	users[userIdx] = *user

	// get max purchase id
	var id int
	for _, purchase := range purchases {
		id = max(purchase.purchaseId, id)
	}

	// add purchase
	purchases = append(purchases, Purchase{
		id + 1,
		userId,
		itemId,
		time.Now(),
	})

	return nil
}

func (t TestDB) UpdateLastLogin(userId int) {
	for i, user := range users {
		if user.userId == userId {
			user.lastLogin = time.Now()
			users[i] = user
			break
		}
	}
}
func (t TestDB) Close() error {
	return nil
}

func NewTestEnv() *Env {
	return &Env{
		logger: log.New(io.Discard, "", 0),
		db:     TestDB{},
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	env := NewTestEnv()

	t.Run("LoginNoBasicAuth", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/api/login", nil)
		env.Login(recorder, request)
		result := recorder.Result()
		if result.StatusCode != http.StatusBadRequest {
			t.Errorf("bad status code for login with no credentials, expected %v, got %v", http.StatusBadRequest, result.StatusCode)
		}
	})

	t.Run("LoginInvalidBasicAuth", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/api/login", nil)
		request.SetBasicAuth("invalid", "invalid")
		env.Login(recorder, request)
		result := recorder.Result()
		if result.StatusCode != http.StatusUnauthorized {
			t.Errorf("bad status code for login with invalid credentials, expected %v, got %v", http.StatusUnauthorized, result.StatusCode)
		}
	})

	t.Run("LoginValid", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/api/login", nil)
		request.SetBasicAuth("test_user", "password")
		env.Login(recorder, request)
		result := recorder.Result()
		if result.StatusCode != http.StatusOK {
			t.Fatalf("bad status code for login with valid credentials, expected %v, got %v", http.StatusOK, result.StatusCode)
		}

		var session *http.Cookie
		var csrf *http.Cookie
		for _, cookie := range result.Cookies() {
			switch cookie.Name {
			case "session_id":
				session = cookie
			case "csrf_token":
				csrf = cookie
			}
		}
		if session == nil {
			t.Error("no session cookie")
		}
		if csrf == nil {
			t.Error("no csrf cookie")
		}
	})
}

func TestRegister(t *testing.T) {
	t.Parallel()

	env := NewTestEnv()

	t.Run("RegisterNoBasicAuth", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/api/register", nil)
		env.Register(recorder, request)
		result := recorder.Result()
		if result.StatusCode != http.StatusBadRequest {
			t.Errorf("bad status code for registration with no credentials, expected %v, got %v", http.StatusBadRequest, result.StatusCode)
		}
	})

	t.Run("RegisterExistingUsername", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/api/register", nil)
		request.SetBasicAuth("test_user", "password")
		env.Register(recorder, request)
		result := recorder.Result()
		if result.StatusCode != http.StatusConflict {
			t.Errorf("bad status code for registration with existing username, expected %v, got %v", http.StatusConflict, result.StatusCode)
		}
	})

	t.Run("RegisterValidUser", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/api/register", nil)
		request.SetBasicAuth("valid_user", "password")
		env.Register(recorder, request)
		result := recorder.Result()
		if result.StatusCode != http.StatusCreated {
			t.Errorf("bad status code for registration with valid user, expected %v, got %v", http.StatusCreated, result.StatusCode)
		}
	})
}
