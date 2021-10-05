package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/storage"
)

func TestGetPhonesByEmailPrefix(t *testing.T) {
	cases := []struct {
		EmailPrefix    string
		ExpectedPhones []*storage.FoundPhone
		MockErr        error
		ExpectedErr    error
	}{
		{
			EmailPrefix: "aliddl",
			ExpectedPhones: []*storage.FoundPhone{
				{
					FirstName: "Alice",
					LastName:  "Liddell",
					Phone:     "+12345",
				},
			},
			MockErr:     nil,
			ExpectedErr: nil,
		},
		{
			EmailPrefix: "ALiddl",
			ExpectedPhones: []*storage.FoundPhone{
				{
					FirstName: "Alice",
					LastName:  "Liddell",
					Phone:     "+12345",
				},
			},
			MockErr:     nil,
			ExpectedErr: nil,
		},
		{
			EmailPrefix:    "",
			ExpectedPhones: nil,
			ExpectedErr:    ErrIncorrectEmailPrefix,
		},
		{
			EmailPrefix:    "aliddl",
			ExpectedPhones: nil,
			MockErr:        fmt.Errorf("some simple err"),
			ExpectedErr:    ErrDBRequestFailed,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("test case #%d", i), func(t *testing.T) {
			mock := &dbMock{
				t:              t,
				expectedPrefix: tc.EmailPrefix,
				expectedError:  tc.MockErr,
				phonesToReturn: tc.ExpectedPhones,
			}
			actualFoundPhones, actualErr := GetPhonesByEmailPrefix(mock, tc.EmailPrefix)
			if t.Failed() {
				return
			}
			if err := compareErrs(tc.ExpectedErr, actualErr); err != nil {
				t.Error(err)
				return
			}
			if len(actualFoundPhones) != len(tc.ExpectedPhones) {
				t.Errorf("expected phones len %d, got %d", len(tc.ExpectedPhones), len(actualFoundPhones))
				return
			}
			for i, expectedPhone := range tc.ExpectedPhones {
				actualPhone := *actualFoundPhones[i]
				if *expectedPhone != actualPhone {
					t.Errorf("phone %d: expected: %v, got: %v", i, *expectedPhone, actualPhone)
					return
				}
			}
		})
	}
}

func compareErrs(expectedErr error, actualErr error) error {
	if expectedErr == nil && actualErr == nil {
		return nil
	}
	if actualErr == nil {
		return fmt.Errorf("expected an error \"%v\", got nil", expectedErr)
	}
	if !errors.Is(actualErr, expectedErr) {
		return fmt.Errorf("expected error \"%v\" and actual error \"%v\" are different", expectedErr, actualErr)
	}
	return nil
}

type dbMock struct {
	t              *testing.T
	expectedPrefix string
	expectedError  error
	phonesToReturn []*storage.FoundPhone
}

func (db *dbMock) GetPhonesByEmailPrefix(ctx context.Context, prefix string) ([]*storage.FoundPhone, error) {
	if prefix != db.expectedPrefix {
		db.t.Errorf("error in DB mock: expected email prefix: %s, got: %s", db.expectedPrefix, prefix)
		return nil, nil
	}
	return db.phonesToReturn, db.expectedError
}
