package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	emailHint "github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/http"
	"github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/storage"
)

func main() {
	srv, err := initServer()
	if err != nil {
		log.Fatalf("[ERR]: failed to initialize server: %v", err)
	}
	log.Println("Let's Go!")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("[ERR]: %v", err)
	}
}

func initServer() (*http.Server, error) {
	handler, err := registerRoutes()
	if err != nil {
		return nil, fmt.Errorf("failed to register routes: %w", err)
	}
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	return srv, nil
}

func registerRoutes() (http.Handler, error) {
	r := mux.NewRouter()
	r.HandleFunc("/phone/{emailPrefix}", func(w http.ResponseWriter, r *http.Request) {
		emailHint.GetPhonesByEmailPrefix(w, r, mux.Vars(r)["emailPrefix"])
	}).Methods("GET")
	addDBMiddleware, err := createAddDBMiddleware()
	if err != nil {
		return nil, fmt.Errorf("failed to create AddDBMiddleware: %w", err)
	}
	r.Use(addDBMiddleware)
	return r, err
}

func createAddDBMiddleware() (func(next http.Handler) http.Handler, error) {
	connStr, err := getConnString()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string info for DB connection: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			db, err := storage.NewDB(connStr)
			if err != nil {
				log.Println("[ERR]: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer db.Close()
			r = r.WithContext(context.WithValue(r.Context(), storage.ContextKeyDB, db))

			next.ServeHTTP(w, r)
		})
	}, nil
}

const (
	dbConnVarNameHost     = "DB_HOST"
	dbConnVarNamePort     = "DB_PORT"
	dbConnVarNameUser     = "DB_USER"
	dbConnVarNamePassword = "DB_PASSWORD"
	dbConnVarNameDBName   = "DB_NAME"
)

func getConnString() (*storage.ConnString, error) {
	fnLookupVar := func(varName string) (string, error) {
		val, ok := os.LookupEnv(varName)
		if !ok {
			return "", fmt.Errorf("variable %s is not defined", varName)
		}
		return val, nil
	}
	connStr := &storage.ConnString{}
	var err error
	connStr.Host, err = fnLookupVar(dbConnVarNameHost)
	if err != nil {
		return nil, err
	}
	connStr.Port, err = fnLookupVar(dbConnVarNamePort)
	if err != nil {
		return nil, err
	}
	connStr.User, err = fnLookupVar(dbConnVarNameUser)
	if err != nil {
		return nil, err
	}
	connStr.Password, err = fnLookupVar(dbConnVarNamePassword)
	if err != nil {
		return nil, err
	}
	connStr.DBName, err = fnLookupVar(dbConnVarNameDBName)
	if err != nil {
		return nil, err
	}
	return connStr, nil
}
