package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/service"
	"github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/storage"
)

func GetPhonesByEmailPrefix(w http.ResponseWriter, r *http.Request, emailPrefix string) {
	dbIface := r.Context().Value(storage.ContextKeyDB)
	if dbIface == nil {
		log.Println("DB is not found in the request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	db, ok := dbIface.(storage.DB)
	if !ok {
		log.Println("DB in the request context is not of type storage.DB")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	phones, err := service.GetPhonesByEmailPrefix(db, emailPrefix)
	if err != nil {
		log.Println(err)
		if errors.Is(err, service.ErrIncorrectEmailPrefix) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDBRequestFailed) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(phones)
	if err != nil {
		log.Printf("failed to serialize the phones list to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("failed to write the phones list as a response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
