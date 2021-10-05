package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	emailHint "github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/http"
)

func main() {
	srv := initServer()
	log.Println("Let's Go!")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("[ERR]: %v", err)
	}
}

func initServer() *http.Server {
	handler := registerRoutes()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	return srv
}

func registerRoutes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/phone/{emailPrefix}", emailHint.GetPhonesByEmailPrefix).
		Methods("GET")
	return r
}
