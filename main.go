package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var err error
	var env *Env

	env, err = NewEnv()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer env.db.Close()

	http.HandleFunc("GET   /api/items", env.PanicMiddleware(env.LogMiddleware(env.AuthMiddleware(env.Items))))
	http.HandleFunc("GET   /api/purchases", env.PanicMiddleware(env.LogMiddleware(env.AuthMiddleware(env.Purchases))))
	http.HandleFunc("GET   /api/balance", env.PanicMiddleware(env.LogMiddleware(env.AuthMiddleware(env.Balance))))
	http.HandleFunc("PATCH /api/deposit", env.PanicMiddleware(env.LogMiddleware(env.AuthMiddleware(env.Deposit))))
	http.HandleFunc("POST  /api/purchase", env.PanicMiddleware(env.LogMiddleware(env.AuthMiddleware(env.Purchase))))
	http.HandleFunc("POST  /api/register", env.PanicMiddleware(env.LogMiddleware(env.Register)))
	http.HandleFunc("POST  /api/login", env.PanicMiddleware(env.LogMiddleware(env.Login)))
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
