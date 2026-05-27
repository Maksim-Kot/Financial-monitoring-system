package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	appErrors "fms-project/internal/infrastructure/errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

//go:embed migrations/*.sql
var fs embed.FS

type ClientParams struct {
	DSN string
}

type Client struct {
	db  *bun.DB
	dsn string
}

func NewClient(p ClientParams) (*Client, error) {
	db, err := connect(p.DSN)
	if err != nil {
		return nil, err
	}
	return &Client{
		db:  db,
		dsn: p.DSN,
	}, nil
}

func (c *Client) Migrate(_ context.Context) error {
	m, err := c.getMigrateInstance()
	if err != nil {
		return appErrors.Wrap(err, "postgres-migrate", "failed to create migrate instance")
	}
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	if err != nil {
		return appErrors.Wrap(err, "postgres-migrate", "failed to migrate db")
	}
	return nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func connect(dsn string) (*bun.DB, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	db := bun.NewDB(sqlDB, pgdialect.New())

	if err := db.Ping(); err != nil {
		return nil, appErrors.Wrap(err, "postgres-client", "failed to connect to postgres")
	}

	return db, nil
}

func (c *Client) getMigrateInstance() (*migrate.Migrate, error) {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create iofs instance: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, c.dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	return m, nil
}
