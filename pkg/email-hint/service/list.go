package service

import (
	"github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/storage"
)

func GetPhonesByEmailPrefix(db storage.DB, emailPrefix string) ([]*storage.FoundPhone, error) {
	return nil, nil
}
