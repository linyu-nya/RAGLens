package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linyu-nya/RAGLens/raglens-server/internal/config"
)

func NewPoolConfig(cfg config.DatabaseConfig) (*pgxpool.Config, error) {
	if cfg.URL == "" {
		return nil, errors.New("database url is required")
	}
	return pgxpool.ParseConfig(cfg.URL)
}
