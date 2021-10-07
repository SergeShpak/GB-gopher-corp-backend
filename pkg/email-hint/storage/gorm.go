package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Employee struct {
	ID         int       `gorm:"column:id"`
	FirstName  string    `gorm:"column:first_name"`
	LastName   string    `gorm:"column:last_name"`
	Salary     string    `gorm:"column:salary"`
	ManagerID  int       `gorm:"column:manager_id"`
	Department int       `gorm:"column:department"`
	Position   int       `gorm:"column:position"`
	EntryAt    time.Time `gorm:"column:entry_at"`
	Phone      string    `gorm:"column:phone"`
	Email      string    `gorm:"column:email"`
}

type gormDB struct {
	db *gorm.DB
}

func newGormDB(c *ConnString) (*gormDB, error) {
	db, err := gorm.Open(postgres.Open(composeGormDSN(c)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open Gorm connection: %w", err)
	}
	return &gormDB{
		db: db,
	}, nil
}

func (g *gormDB) GetPhonesByEmailPrefix(ctx context.Context, prefix string) ([]*FoundPhone, error) {
	var emps []Employee
	req := g.db.
		Select("first_name", "last_name", "phone", "email").
		Where("email LIKE ?", prefix+"%").
		Find(&emps)
	if err := req.Error; err != nil {
		return nil, fmt.Errorf("failed to query phones by email: %w", err)
	}
	phones := make([]*FoundPhone, len(emps))
	for i, e := range emps {
		p := &FoundPhone{
			FirstName: e.FirstName,
			LastName:  e.LastName,
			Phone:     e.Phone,
			Email:     e.Email,
		}
		phones[i] = p
	}
	return phones, nil
}

func (g *gormDB) Close() {}

func composeGormDSN(c *ConnString) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		c.Host,
		c.User,
		c.Password,
		c.DBName,
		c.Port,
	)
}
