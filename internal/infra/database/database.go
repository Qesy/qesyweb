package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Qesy/qesysql"
	"github.com/Qesy/QesyWeb/config"
)

func Open(ctx context.Context, cfg config.MySQLConfig) (*qesysql.DB, error) {
	if !cfg.Enabled { return nil, nil }
	if cfg.DSN == "" { return nil, fmt.Errorf("MYSQL_DSN is required when MYSQL_ENABLED=true") }
	return qesysql.Open(ctx, cfg.Driver, cfg.DSN, qesysql.Config{MaxOpenConns: cfg.MaxOpen, MaxIdleConns: cfg.MaxIdle, ConnMaxLifetime: 30 * time.Minute, ConnMaxIdleTime: 5 * time.Minute})
}
