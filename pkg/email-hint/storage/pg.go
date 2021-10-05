package storage

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type ContextKey int

const ContextKeyDB ContextKey = iota + 1

var (
	db    *pgxpool.Pool
	dbMux *sync.Mutex
)

type FoundPhone struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `phone:"phone"`
}

type DB interface {
	GetPhonesByEmailPrefix(ctx context.Context, prefix string) ([]*FoundPhone, error)
}

func NewDB() (DB, error) {
	pool, err := getConn()
	if err != nil {
		return nil, fmt.Errorf("failed to get a connection pool: %w", err)
	}
	return &conn{
		db: pool,
	}, nil
}

type conn struct {
	db *pgxpool.Pool
}

func (c *conn) GetPhonesByEmailPrefix(ctx context.Context, prefix string) ([]*FoundPhone, error) {
	return nil, nil
}

func getConn() (*pgxpool.Pool, error) {
	dbMux.Lock()
	defer dbMux.Unlock()
	if db != nil {
		return db, nil
	}

	var err error
	db, err = initPGXPool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a PGX pool: %w", err)
	}
	return db, nil
}

func initPGXPool() (*pgxpool.Pool, error) {
	connStr, err := composeConnectionString()
	if err != nil {
		return nil, fmt.Errorf("failed to compose the connection string: %w", err)
	}
	cfg, err := getPGXPoolConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get the PGX pool config: %w", err)
	}
	db, err = pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the postgres DB using a PGX connection pool: %w", err)
	}
	return db, nil
}

func getPGXPoolConfig(connStr string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create the PGX pool config from connection string: %w", err)
	}
	cfg.ConnConfig.ConnectTimeout = time.Second * 1
	return cfg, nil
}

const (
	dbConnVarNameHost     = "DB_HOST"
	dbConnVarNamePort     = "DB_PORT"
	dbConnVarNameUser     = "DB_USER"
	dbConnVarNamePassword = "DB_PASSWORD"
	dbConnVarNameDBName   = "DB_NAME"
)

func composeConnectionString() (string, error) {
	fnLookupVar := func(varName string) (string, error) {
		val, ok := os.LookupEnv(varName)
		if !ok {
			return "", fmt.Errorf("variable %s is not defined", varName)
		}
		return val, nil
	}
	host, err := fnLookupVar(dbConnVarNameHost)
	if err != nil {
		return "", nil
	}
	port, err := fnLookupVar(dbConnVarNamePort)
	if err != nil {
		return "", nil
	}
	user, err := fnLookupVar(dbConnVarNameUser)
	if err != nil {
		return "", nil
	}
	password, err := fnLookupVar(dbConnVarNamePassword)
	if err != nil {
		return "", nil
	}
	dbName, err := fnLookupVar(dbConnVarNameDBName)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", host, port, user, password, dbName), nil
}
