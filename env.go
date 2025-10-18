package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type CtxKey uint

const (
	CtxUserId CtxKey = iota
)

type Env struct {
	logger *log.Logger
	db     DB
}

func NewEnv() (*Env, error) {
	sqlDb, err := NewSqlDB()
	if err != nil {
		return nil, err
	}

	return &Env{
		logger: log.Default(),
		db:     sqlDb,
	}, err
}

func (env *Env) Balance(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(CtxUserId).(int)
	if !ok {
		env.logger.Println("context does not include userId for protected endpoint")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Get balance
	balance, err := env.db.Balance(userId)
	if err != nil {
		env.logger.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "Balance:", convertMoneyPrintable(balance))
}

func (env *Env) Purchase(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(CtxUserId).(int)
	if !ok {
		env.logger.Println("context does not include userId for protected endpoint")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Get item id
	itemId, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Attempt purchase
	if err := env.db.Purchase(userId, itemId); err != nil {
		var statusCode int
		switch err {
		case ErrInsufficientFunds:
			statusCode = http.StatusForbidden
			fmt.Fprintln(w, "Insufficient funds")
		default:
			env.logger.Println(err.Error())
			statusCode = http.StatusInternalServerError
		}
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	fmt.Fprintln(w, "Success")
}

func (env *Env) Deposit(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(CtxUserId).(int)
	if !ok {
		env.logger.Println("context does not include userId for protected endpoint")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var err error

	// Parse deposit amount (conver to internal integer representation)
	var depositAmountFloat float64
	depositAmountFloat, err = strconv.ParseFloat(r.FormValue("amount"), 64)
	depositAmount := int(depositAmountFloat * 100)
	if err != nil || depositAmount <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Deposit money
	balance, err := env.db.Deposit(userId, depositAmount)
	if err != nil {
		env.logger.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "New balance:", convertMoneyPrintable(balance))
}

func (env *Env) Items(w http.ResponseWriter, r *http.Request) {
	// Get items
	items, err := env.db.Items()
	if err != nil {
		env.logger.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Print items
	for _, item := range items {
		fmt.Fprintln(w, item)
	}
}

func (env *Env) Purchases(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(CtxUserId).(int)
	if !ok {
		env.logger.Println("context does not include userId for protected endpoint")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Get purchases
	purchases, err := env.db.Purchases(userId)
	if err != nil {
		env.logger.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Print purchases
	for _, purchase := range purchases {
		fmt.Fprintln(w, purchase)
	}
}

func (env *Env) Register(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Check username isn't taken
	_, err := env.db.GetUserFromUsername(username)
	if err == nil {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		fmt.Fprintln(w, "Account with username already exists")
		return
	}

	// Hash password
	passwordHash, err := hashPassword(password)
	if err != nil {
		env.logger.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Register account
	user, err := env.db.Register(username, passwordHash)
	if err != nil {
		env.logger.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Log and return
	env.logger.Printf("%q registered\n", user.username)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Success")
}

func (env *Env) Login(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get user
	user, err := env.db.GetUserFromUsername(username)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Validate password
	if !checkPasswordHash(password, user.passwordHash) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Parse IP addr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		env.logger.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Create session
	session, err := env.db.CreateSession(user, host)
	if err != nil {
		env.logger.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Send session cookies
	expiresAt := time.Now().Add(time.Hour * 24)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.sessionId,
		Expires:  expiresAt,
		HttpOnly: true, // prevent client side js from reading
		SameSite: http.SameSiteLaxMode,
		// Secure: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    session.csrfToken,
		Expires:  expiresAt,
		HttpOnly: false, // need js to read to put in header
		SameSite: http.SameSiteLaxMode,
		// Secure: true,
	})

	// Update last login
	env.db.UpdateLastLogin(session.userId)

	fmt.Fprintln(w, "Success")
}

func (env *Env) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session id cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Get session from database
		session, err := env.db.GetSession(cookie.Value)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Validate session against csrf token in header
		if csrfToken := r.Header.Get("X-CSRF-Token"); csrfToken != session.csrfToken {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		env.db.UpdateLastLogin(session.userId)

		// Add userId to context
		*r = *r.WithContext(context.WithValue(r.Context(), CtxUserId, session.userId))

		next(w, r)
	}
}

func (env *Env) LogMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			env.logger.Println(r.RemoteAddr, r.Method, r.URL, time.Since(start))
		}()
		next(w, r)
	}
}

func (env *Env) PanicMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				env.logger.Printf("Recovered from panic: %v\n", r)
				debug.PrintStack()
			}
		}()
		next(w, r)
	}
}
