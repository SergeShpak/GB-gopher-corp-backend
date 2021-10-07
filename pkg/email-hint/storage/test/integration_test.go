// +build integration_tests

package storage

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/SergeyShpak/gopher-corp-backend/pkg/email-hint/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	DB_HOST     = "127.0.0.1"
	DB_USER     = "gopher"
	DB_PASSWORD = "P@ssw0rd"
	DB_NAME     = "gopher_corp"
)

var DB_PORT = ""

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	setupResult, err := setup()
	if err != nil {
		log.Println("setup err: ", err)
		return -1
	}
	defer teardown(setupResult)
	return m.Run()
}

type teardownPack struct {
	OldEnvVars map[string]string
}

const dataDir = "data"

type setupResult struct {
	Pool              *dockertest.Pool
	PostgresContainer *dockertest.Resource
}

const dockerMaxWait = time.Second * 5

func setup() (r *setupResult, err error) {
	testFileDir, err := getTestFileDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get the script dir: %w", err)
	}
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to create a new docketest pool: %w", err)
	}
	pool.MaxWait = dockerMaxWait

	postgresContainer, err := runPostgresContainer(pool, testFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run the Postgres container: %w", err)
	}
	defer func() {
		if err != nil {
			if err := pool.Purge(postgresContainer); err != nil {
				log.Println("failed to purge the postgres container: %w", err)
			}
		}
	}()

	migrationContainer, err := runMigrationContainer(pool, testFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run the migration container: %w", err)
	}

	defer func() {
		if err := pool.Purge(migrationContainer); err != nil {
			err = fmt.Errorf("failed to purge the migration container: %w", err)
		}
	}()

	if err := pool.Retry(func() error {
		err := prepopulateDB(testFileDir)
		if err != nil {
			log.Printf("populate DB err: %v", err)
		}
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to prepopulate the DB: %w", err)
	}

	return &setupResult{
		Pool:              pool,
		PostgresContainer: postgresContainer,
	}, nil
}

func getTestFileDir() (string, error) {
	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get the caller info")
	}
	fileDir := filepath.Dir(fileName)
	dir, err := filepath.Abs(fileDir)
	if err != nil {
		return "", fmt.Errorf("failed to get the absolute path to the directory %s: %w", dir, err)
	}
	return fileDir, nil
}

func runPostgresContainer(pool *dockertest.Pool, testFileDir string) (*dockertest.Resource, error) {
	postgresContainer, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "14.0",
			Env: []string{
				"POSTGRES_PASSWORD=P@ssw0rd",
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = false
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
			config.Mounts = []docker.HostMount{
				{
					Target: "/docker-entrypoint-initdb.d",
					Source: filepath.Join(testFileDir, "init"),
					Type:   "bind",
				},
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start the postgres docker container: %w", err)
	}
	postgresContainer.Expire(120)

	DB_PORT = postgresContainer.GetPort("5432/tcp")

	// Wait for the DB to start
	if err := pool.Retry(func() error {
		db, err := getDBConnector()
		if err != nil {
			return fmt.Errorf("failed to get a DB connector: %w", err)
		}
		return db.Ping(context.Background())
	}); err != nil {
		pool.Purge(postgresContainer)
		return nil, fmt.Errorf("failed to ping the created DB: %w", err)
	}
	return postgresContainer, nil
}

func runMigrationContainer(pool *dockertest.Pool, testFileDir string) (*dockertest.Resource, error) {
	migrationsDir, err := filepath.Abs(filepath.Join(testFileDir, "../../../../migrations"))
	if err != nil {
		return nil, fmt.Errorf("failed to get the absolute path of the migrations dir: %w", err)
	}
	migrationContainer, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "migrate/migrate",
			Tag:        "v4.15.0",
			Cmd: []string{
				"-path=/migrations",
				fmt.Sprintf(
					"-database=%s",
					composeConnectionString(),
				),
				"up",
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = false
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
			config.Mounts = []docker.HostMount{
				{
					Target: "/migrations",
					Source: migrationsDir,
					Type:   "bind",
				},
			}
			config.NetworkMode = "host"
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start the migration container: %w", err)
	}

	return migrationContainer, err
}

func prepopulateDB(testFileDir string) error {
	prepopulateScriptPath := filepath.Join(testFileDir, "prepopulate_db.sql")
	scriptBytes, err := os.ReadFile(prepopulateScriptPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", prepopulateScriptPath, err)
	}
	conn, err := getDBConnector()
	if err != nil {
		return fmt.Errorf("failed to get a DB connector: %w", err)
	}
	if _, err := conn.Exec(context.Background(), string(scriptBytes)); err != nil {
		return fmt.Errorf("failed to execute the prepopulate script: %w", err)
	}
	return nil
}

func teardown(r *setupResult) {
	if err := r.Pool.Purge(r.PostgresContainer); err != nil {
		log.Printf("failed to purge the Postgres container: %v", err)
	}
}

func TestGetPhonesByEmailPrefix(t *testing.T) {
	conn, err := getDBConnector()
	if err != nil {
		t.Fatalf("failed to get a connector to the DB: %v", err)
	}

	emailTestPrefix := "test_get_phones_by_email_prefix"
	employees := []storage.FoundPhone{
		{
			FirstName: "Dale",
			LastName:  "Cooper",
			Phone:     "+72345",
			Email:     "test_get_phones_by_email_prefix_dcooper@gopher_corp.com",
		},
		{
			FirstName: "Bobby",
			LastName:  "Briggs",
			Phone:     "+73456",
			Email:     "test_get_phones_by_email_prefix_bbriggs@gopher_corp.com",
		},
		{
			FirstName: "Audrey",
			LastName:  "Horne",
			Phone:     "+74567",
			Email:     "test_get_phones_by_email_prefix_ahorne@gopher_corp.com",
		},
	}
	batch := &pgx.Batch{}
	const query = `INSERT INTO employees (first_name, last_name, phone, email, salary, manager_id, department, position)
		VALUES(
			$1, $2, $3, $4, 
			45000,
			(SELECT id FROM employees WHERE first_name = 'Bob' AND last_name = 'Morane' LIMIT 1),
			(SELECT id FROM departments WHERE name = 'R&D'),
			(SELECT id FROM positions WHERE title = 'Backend Dev')
		)`
	for _, e := range employees {
		batch.Queue(
			query,
			e.FirstName,
			e.LastName,
			e.Phone,
			fmt.Sprintf(
				"%s_%s%s@gopher_corp.com",
				emailTestPrefix,
				string(strings.ToLower(e.FirstName)[0]),
				strings.ToLower(e.LastName),
			),
		)
	}
	if _, err := conn.SendBatch(context.Background(), batch).Exec(); err != nil {
		t.Fatalf("failed to create DB data: %v", err)
	}

	db, err := storage.NewDB(getConnectionString())
	if err != nil {
		t.Fatalf("failed to create a DB object: %v", err)
	}
	phones, err := db.GetPhonesByEmailPrefix(context.Background(), emailTestPrefix)
	if err != nil {
		t.Fatalf("GetPhonesByEmailPrefix failed: %v", err)
	}
	if len(phones) != len(employees) {
		t.Fatalf("expected length of found phones to be %d, got %d", len(employees), len(phones))
	}
	sort.Slice(employees, func(i int, j int) bool {
		return employees[i].FirstName < employees[j].FirstName
	})
	sort.Slice(phones, func(i int, j int) bool {
		return phones[i].FirstName < phones[j].FirstName
	})
	for i, e := range employees {
		p := *phones[i]
		if e != p {
			t.Fatalf("expected object %v is not equal to the actual object %v", e, p)
		}
	}
}

func getDBConnector() (*pgxpool.Pool, error) {
	log.Println(composeConnectionString())
	cfg, err := pgxpool.ParseConfig(composeConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to create the PGX pool config from connection string: %w", err)
	}
	cfg.ConnConfig.ConnectTimeout = time.Second * 1
	db, err := pgxpool.ConnectConfig(context.Background(), cfg)
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

func getConnectionString() *storage.ConnString {
	return &storage.ConnString{
		Host:     DB_HOST,
		Port:     DB_PORT,
		User:     DB_USER,
		Password: DB_PASSWORD,
		DBName:   DB_NAME,
	}
}

func composeConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", DB_USER, url.QueryEscape(DB_PASSWORD), DB_HOST, DB_PORT, DB_NAME)
}
