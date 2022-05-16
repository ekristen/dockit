package db

import (
	"context"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New creates a new database connection
func New(ctx context.Context, dialect string, dsn string, config *gorm.Config) (db *gorm.DB, err error) {
	if config == nil {
		config = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	}

	if dialect == "sqlite" {
		config.DisableForeignKeyConstraintWhenMigrating = true
		db, err = NewSQLite(dsn, config)
	} else if dialect == "mysql" {
		db, err = NewMySQL(dsn, config)
	} else {
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}

	db = db.WithContext(ctx)

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&User{},
		&Group{},
		&Permission{},
		&Token{},
		&PKI{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

// NewMySQL Creates a new MySQL Database Connection
func NewMySQL(dsn string, config *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), config)
}

// NewSQLite --
func NewSQLite(file string, config *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(file), config)
}
