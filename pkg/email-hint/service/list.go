package service

import (
	"context"
	"fmt"

	"github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/storage"
)

var (
	ErrIncorrectEmailPrefix = fmt.Errorf("got an incorrect email prefix")
	ErrDBRequestFailed      = fmt.Errorf("a request to DB failed")
)

func GetPhonesByEmailPrefix(db storage.DB, emailPrefix string) ([]*storage.FoundPhone, error) {
	if len(emailPrefix) == 0 {
		return nil, fmt.Errorf("%w: the passed prefix is empty", ErrIncorrectEmailPrefix)
	}
	phones, err := db.GetPhonesByEmailPrefix(context.Background(), emailPrefix)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get phones by email prefix: %v", ErrDBRequestFailed, err)
	}
	return phones, nil
}
