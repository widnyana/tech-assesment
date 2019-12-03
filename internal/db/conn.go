package db

import (
	"database/sql"
	_ "github.com/Go-SQL-Driver/MySQL"
	"kumparan/internal/config"
	"time"
)

var (
	dbCon *sql.DB
)

func InitializeDBConnection(cfg config.DBConf) error {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(cfg.MaxidleCon)
	db.SetMaxOpenConns(cfg.MaxOpenCon)
	db.SetConnMaxLifetime(time.Minute * cfg.Lifetime)
	dbCon = db

	return nil
}

func GetDB() *sql.DB {
	return dbCon
}
