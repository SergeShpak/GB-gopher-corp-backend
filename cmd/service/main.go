package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	emailHint "github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/http"
	"github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/storage"
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
	r.HandleFunc("/phone/{emailPrefix}", func(w http.ResponseWriter, r *http.Request) {
		emailHint.GetPhonesByEmailPrefix(w, r, mux.Vars(r)["emailPrefix"])
	}).Methods("GET")
	r.Use(addDBMiddleware)
	return r
}

func addDBMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := storage.NewDB()
		if err != nil {
			log.Println("[ERR]: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), storage.ContextKeyDB, db))
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
