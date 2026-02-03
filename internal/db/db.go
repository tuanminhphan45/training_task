package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"task1/internal/config"
	"task1/internal/model"
)

func Init(cfg *config.DBConfig) *bun.DB {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqldb.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	db := bun.NewDB(sqldb, pgdialect.New())
	db.NewCreateTable().Model((*model.Hash)(nil)).IfNotExists().Exec(context.Background())

	return db
}
