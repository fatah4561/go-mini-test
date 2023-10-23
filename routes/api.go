package routes

import (
	"log"
	"net/http"

	"go-mini-test/auth"
	"go-mini-test/repositories"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func HandleRequest() {
	// acts as route

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/api/login", auth.Login).Methods("POST")
	myRouter.HandleFunc("/api/user", auth.JWTMiddleware(repositories.CreateUser)).Methods("POST")        // create
	myRouter.HandleFunc("/api/user", auth.JWTMiddleware(repositories.GetUsers))                          // all
	myRouter.HandleFunc("/api/user/{id}", auth.JWTMiddleware(repositories.UpdateUser)).Methods("PATCH")  // update
	myRouter.HandleFunc("/api/user/{id}", auth.JWTMiddleware(repositories.DeleteUser)).Methods("DELETE") // delete
	myRouter.HandleFunc("/api/user/{id}", auth.JWTMiddleware(repositories.GetUserDetail))                // detail

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}
